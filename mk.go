package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mattn/go-isatty"
)

var (
	// True if messages should be printed with fancy colors.
	// By default, if the output stream is not the terminal, colors are disabled.
	color bool

	// Default shell to use if none specified via $shell.
	defaultShell string

	// Do not drop shell arguments when calling with no further arguments
	// This works around `sh -c commands...` being a thing, but allows the `rc -v commands...` argument-less flags
	dontDropArgs bool

	// True if we are ignoring timestamps and rebuilding everything.
	rebuildall bool = false

	// Set of targets for which we are forcing rebuild
	rebuildtargets map[string]bool = make(map[string]bool)

	// Lock on standard out, messages don't get interleaved too much.
	mkMsgMutex sync.Mutex

	// sched controls parallel recipe execution.
	sched scheduler

	// The maximum number of times an rule may be applied.
	// This limits recursion of both meta- and non-meta-rules!
	// Maybe, this shouldn't affect meta-rules?!
	maxRuleCnt int = 1

	// Keep going after errors (-k flag).
	keepgoing bool

	// Touch targets instead of executing recipes (-t flag).
	touchmode bool

	// Pretend this target was recently modified (-w flag).
	pretendModified string

	// Force rebuild of missing intermediates (-i flag).
	forceIntermediate bool

	// Print explanation of staleness decisions (-e flag).
	explain bool

	// Set when any recipe fails; checked to stop new recipes when -k is not set.
	buildFailed atomic.Bool
)

// scheduler controls parallel recipe execution, limiting the number of
// concurrent subprocesses.
type scheduler struct {
	allowed   int
	running   int
	cond      *sync.Cond
	exclusive sync.Mutex
}

// Wait until there is an available subprocess slot.
// Returns the 0-based slot number assigned to this job.
func (s *scheduler) reserve() int {
	s.cond.L.Lock()
	for s.running >= s.allowed {
		s.cond.Wait()
	}
	slot := s.running
	s.running++
	s.cond.L.Unlock()
	return slot
}

// Free up another subprocess to run.
func (s *scheduler) finish() {
	s.cond.L.Lock()
	s.running--
	s.cond.Signal()
	s.cond.L.Unlock()
}

// Acquire exclusive access, waiting for all running subprocesses to finish.
func (s *scheduler) reserveExclusive() {
	s.exclusive.Lock()
	stolenSubprocs := 0
	s.cond.L.Lock()
	stolenSubprocs = s.allowed - s.running
	s.running = s.allowed
	for stolenSubprocs < s.allowed {
		s.cond.Wait()
		stolenSubprocs += s.allowed - s.running
		s.running = s.allowed
	}
}

func (s *scheduler) finishExclusive() {
	s.running = 0
	s.cond.Broadcast()
	s.cond.L.Unlock()
	s.exclusive.Unlock()
}

// Ansi color codes.
const (
	ansiTermDefault   = "\033[0m"
	// ansiTermBlack   = "\033[30m"
	ansiTermRed       = "\033[31m"
	// ansiTermGreen  = "\033[32m"
	// ansiTermYellow = "\033[33m"
	ansiTermBlue      = "\033[34m"
	// ansiTermMagenta = "\033[35m"
	ansiTermBright    = "\033[1m"
	ansiTermUnderline = "\033[4m"
)

// Build a node's prereqs. Block until completed.
func mkNodePrereqs(g *graph, prereqs []*node, vars map[string][]string, unexportedVars map[string]bool, dryrun bool, required bool) nodeStatus {
	// When limited to a single process, build sequentially to preserve order.
	if sched.allowed == 1 {
		status := nodeStatusDone
		for i := range prereqs {
			prereqs[i].mutex.Lock()
			switch prereqs[i].status {
			case nodeStatusReady, nodeStatusNop:
				prereqs[i].mutex.Unlock()
				mkNode(g, prereqs[i], vars, unexportedVars, dryrun, required)
			default:
				prereqs[i].mutex.Unlock()
			}
			if prereqs[i].status == nodeStatusFailed {
				status = nodeStatusFailed
			}
		}
		return status
	}

	prereqstat := make(chan nodeStatus)
	pending := 0

	// build prereqs that need building
	for i := range prereqs {
		prereqs[i].mutex.Lock()
		switch prereqs[i].status {
		case nodeStatusReady, nodeStatusNop:
			go mkNode(g, prereqs[i], vars, unexportedVars, dryrun, required)
			fallthrough
		case nodeStatusStarted:
			prereqs[i].listeners = append(prereqs[i].listeners, prereqstat)
			pending++
		}
		prereqs[i].mutex.Unlock()
	}

	// wait until all the prereqs are built
	status := nodeStatusDone
	for pending > 0 {
		s := <-prereqstat
		pending--
		if s == nodeStatusFailed {
			status = nodeStatusFailed
		}
	}
	return status
}

// Build a target in the graph.
//
// This selects an appropriate rule (edge) and builds all prerequisites
// concurrently.
//
// Args:
//
//	g: Graph in which the node lives.
//	u: Node to (possibly) build.
//	dryrun: Don't actually build anything, just pretend.
//	required: Avoid building this node, unless its prereqs are out of date.
func mkNode(g *graph, u *node, vars map[string][]string, unexportedVars map[string]bool, dryrun bool, required bool) {
	// try to claim on this node
	u.mutex.Lock()
	if u.status != nodeStatusReady && u.status != nodeStatusNop {
		u.mutex.Unlock()
		return
	} else {
		u.status = nodeStatusStarted
	}
	u.mutex.Unlock()

	// when finished, notify the listeners
	finalstatus := nodeStatusDone
	defer func() {
		u.mutex.Lock()
		u.status = finalstatus
		for i := range u.listeners {
			u.listeners[i] <- u.status
		}
		u.listeners = u.listeners[0:0]
		u.mutex.Unlock()
	}()

	// there's no rules.
	if len(u.prereqs) == 0 {
		if !(u.r != nil && (u.r.attributes.virtual || u.r.attributes.forcedTimestamp)) && !u.exists {
			wd, _ := os.Getwd()
			mkError(fmt.Sprintf("don't know how to make %s in %s\n", u.name, wd))
		}
		finalstatus = nodeStatusNop
		return
	}

	// there should otherwise be exactly one edge with an associated rule
	prereqs := make([]*node, 0)
	var e *edge = nil
	for i := range u.prereqs {
		if u.prereqs[i].r != nil {
			e = u.prereqs[i]
		}
		if u.prereqs[i].v != nil {
			prereqs = append(prereqs, u.prereqs[i].v)
		}
	}

	prereqs_required := required && (e.r.attributes.virtual || !u.exists || forceIntermediate)
	if mkNodePrereqs(g, prereqs, vars, unexportedVars, dryrun, prereqs_required) == nodeStatusFailed {
		finalstatus = nodeStatusFailed
	}

	uptodate := true
	if !e.r.attributes.virtual {
		u.updateTimestamp()
		if !u.exists && required {
			if explain {
				fmt.Fprintf(os.Stderr, "mk: %s does not exist\n", u.name)
			}
			uptodate = false
		} else if len(e.r.command) > 0 && (u.exists || required) {
			// P attribute: use custom program for staleness checking.
			for i := range prereqs {
				args := append(append([]string{}, e.r.command[1:]...), u.name, prereqs[i].name)
				_, ok := subprocess(e.r.command[0], args, os.Environ(), "", false)
				if !ok {
					if explain {
						fmt.Fprintf(os.Stderr, "mk: %s out of date via %s (P attribute)\n", u.name, prereqs[i].name)
					}
					uptodate = false
					break
				}
			}
		} else if u.exists || required {
			for i := range prereqs {
				if u.t.Before(prereqs[i].t) {
					if explain {
						fmt.Fprintf(os.Stderr, "mk: %s older than %s\n", u.name, prereqs[i].name)
					}
					uptodate = false
				} else if prereqs[i].status == nodeStatusDone {
					if explain {
						fmt.Fprintf(os.Stderr, "mk: %s stale because %s was rebuilt\n", u.name, prereqs[i].name)
					}
					uptodate = false
				}
			}
		}
	} else {
		if explain && u.name != "" { // skip the root dummy node
			fmt.Fprintf(os.Stderr, "mk: %s is virtual\n", u.name)
		}
		uptodate = false
	}

	_, isrebuildtarget := rebuildtargets[u.name]
	if isrebuildtarget || rebuildall {
		if explain && uptodate {
			fmt.Fprintf(os.Stderr, "mk: %s forced by -a/-w flag\n", u.name)
		}
		uptodate = false
	}

	// make another pass on the prereqs, since we know we need them now
	if !uptodate {
		if mkNodePrereqs(g, prereqs, vars, unexportedVars, dryrun, true) == nodeStatusFailed {
			finalstatus = nodeStatusFailed
		}
	}

	// Without -k, stop building when any recipe has failed.
	if !keepgoing && buildFailed.Load() {
		finalstatus = nodeStatusFailed
	}

	// execute the recipe, unless the prereqs failed
	if !uptodate && finalstatus != nodeStatusFailed && len(e.r.recipe) > 0 {
		if touchmode && !e.r.attributes.virtual {
			// Touch mode: update the target's timestamp without running the recipe.
			now := time.Now()
			if !u.exists {
				f, err := os.Create(u.name)
				if err == nil {
					f.Close()
				}
			} else {
				os.Chtimes(u.name, now, now)
			}
			u.updateTimestamp()
		} else if !touchmode {
			var nproc int
			if e.r.attributes.exclusive {
				sched.reserveExclusive()
				nproc = 0
			} else {
				nproc = sched.reserve()
			}

			if !dorecipe(u.name, u, e, vars, unexportedVars, dryrun, nproc) {
				finalstatus = nodeStatusFailed
				buildFailed.Store(true)
				// D attribute: delete the target file when the recipe fails.
				if e.r.attributes.delFailed {
					os.Remove(u.name)
				}
			}
			// U attribute: force timestamp so dependents see the target as updated
			// even if the recipe didn't modify the file.
			if finalstatus != nodeStatusFailed && e.r.attributes.update {
				u.t = time.Now()
			} else {
				u.updateTimestamp()
			}

			if e.r.attributes.exclusive {
				sched.finishExclusive()
			} else {
				sched.finish()
			}
		}
	} else if finalstatus != nodeStatusFailed {
		if explain && uptodate && !e.r.attributes.virtual {
			fmt.Fprintf(os.Stderr, "mk: %s is up to date\n", u.name)
		}
		finalstatus = nodeStatusNop
	}
}

func mkError(msg string) {
	mkPrintError(msg)
	os.Exit(1)
}

func mkPrintError(msg string) {
	if color {
		os.Stderr.WriteString(ansiTermRed)
	}
	fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	if color {
		os.Stderr.WriteString(ansiTermDefault)
	}
}

func mkPrintRecipe(target string, recipe string, quiet bool) {
	mkMsgMutex.Lock()
	if !color {
		fmt.Printf("%s: ", target)
	} else {
		fmt.Printf("%s%s%s → %s",
			ansiTermBlue+ansiTermBright+ansiTermUnderline, target,
			ansiTermDefault, ansiTermBlue)
	}
	if quiet {
		if !color {
			fmt.Println("...")
		} else {
			fmt.Println("…")
		}
	} else {
		printIndented(os.Stdout, recipe, len(target)+3)
		if len(recipe) == 0 {
			os.Stdout.WriteString("\n")
		}
	}
	if color {
		os.Stdout.WriteString(ansiTermDefault)
	}
	mkMsgMutex.Unlock()
}

func main() {
	var directory string
	var mkfilepath string
	var interactive bool
	var dryrun bool
	var shallowrebuild bool
	var quiet bool
	var dotOutput bool

	flag.StringVar(&directory, "C", "", "directory to change in to")
	flag.StringVar(&mkfilepath, "f", "mkfile", "use the given file as mkfile")
	flag.BoolVar(&dryrun, "n", false, "print commands without actually executing")
	flag.BoolVar(&touchmode, "t", false, "touch targets instead of executing recipes")
	flag.BoolVar(&shallowrebuild, "r", false, "force building of just targets")
	flag.BoolVar(&rebuildall, "a", false, "force building of all dependencies")
	flag.BoolVar(&keepgoing, "k", false, "continue building after errors")
	flag.StringVar(&pretendModified, "w", "", "pretend `target` was recently modified")
	flag.IntVar(&sched.allowed, "p", -1, "maximum number of jobs to execute in parallel")
	flag.IntVar(&maxRuleCnt, "l", 1, "maximum number of times a specific rule can be applied (recursion)")
	flag.BoolVar(&interactive, "I", false, "prompt before executing rules")
	flag.BoolVar(&forceIntermediate, "i", false, "force rebuild of missing intermediates")
	flag.BoolVar(&explain, "e", false, "explain why targets are out of date")
	flag.BoolVar(&quiet, "q", false, "don't print recipes before executing them")
	flag.BoolVar(&dotOutput, "dot", false, "print dependency graph in graphviz dot format and exit")
	flag.BoolVar(&color, "color", isatty.IsTerminal(os.Stdout.Fd()), "turn color on/off")
	flag.StringVar(&defaultShell, "shell", "sh -e", "default shell to use if none are specified via $shell")
	flag.BoolVar(&dontDropArgs, "F", false, "don't drop shell arguments when no further arguments are specified")
	// TODO(rjk): P9P mk command line compatability.
	flag.Parse()

	// Resolve parallelism: -p flag > $NPROC env > NumCPU
	if sched.allowed < 0 {
		if nproc := os.Getenv("NPROC"); nproc != "" {
			if n, err := strconv.Atoi(nproc); err == nil && n > 0 {
				sched.allowed = n
			} else {
				mkError(fmt.Sprintf("invalid $NPROC value: %q", nproc))
			}
		} else {
			sched.allowed = runtime.NumCPU()
		}
	}
	sched.cond = sync.NewCond(&sync.Mutex{})

	if directory != "" {
		err := os.Chdir(directory)
		if err != nil {
			mkError(fmt.Sprintf("changing directory to `%s' failed", directory))
		}
	}

	mkfile, err := os.Open(mkfilepath)
	if err != nil {
		mkError("no mkfile found")
	}
	input, _ := io.ReadAll(mkfile)
	mkfile.Close()

	abspath, _ := filepath.Abs(mkfilepath)

	env := make(map[string][]string)
	for _, elem := range os.Environ() {
		vals := strings.SplitN(elem, "=", 2)
		env[vals[0]] = append(env[vals[0]], vals[1])
	}

	rs := parse(string(input), mkfilepath, abspath, env)
	if quiet {
		for i := range rs.rules {
			rs.rules[i].attributes.quiet = true
		}
	}

	// Separate command-line variable overrides (VAR=value) from targets.
	var targets []string
	for _, arg := range flag.Args() {
		if i := strings.Index(arg, "="); i > 0 && isValidVarName(arg[:i]) {
			rs.vars[arg[:i]] = expand(arg[i+1:], rs.vars, true)
		} else {
			targets = append(targets, arg)
		}
	}

	// build the first non-meta rule in the makefile, if none are given explicitly
	if len(targets) == 0 {
		for i := range rs.rules {
			if !rs.rules[i].ismeta {
				for j := range rs.rules[i].targets {
					targets = append(targets, rs.rules[i].targets[j].spat)
				}
				break
			}
		}
	}

	if len(targets) == 0 {
		fmt.Println("mk: nothing to mk")
		return
	}

	if shallowrebuild {
		for i := range targets {
			rebuildtargets[targets[i]] = true
		}
	}

	// Create a dummy virtual rule that depends on every target
	root := rule{}
	root.targets = []pattern{{false, "", nil}}
	root.attributes = attribSet{false, false, false, false, false, false, false, true, false}
	root.prereqs = targets
	rs.add(root)

	if dotOutput {
		g := buildgraph(rs, "", maxRuleCnt)
		g.visualize(os.Stdout)
		return
	}

	if interactive {
		g := buildgraph(rs, "", maxRuleCnt)
		mkNode(g, g.root, rs.vars, rs.unexportedVars, true, true)
		fmt.Print("Proceed? ")
		in := bufio.NewReader(os.Stdin)
		for {
			c, _, err := in.ReadRune()
			if err != nil {
				return
			} else if strings.ContainsRune(" \n\t\r", c) {
				continue
			} else if c == 'y' {
				break
			} else {
				return
			}
		}
	}

	g := buildgraph(rs, "", maxRuleCnt)

	// -w flag: pretend a target was recently modified.
	if pretendModified != "" {
		if n, ok := g.nodes[pretendModified]; ok {
			n.t = time.Now()
			n.flags |= nodeFlagProbable | nodeFlagForcedTime
		}
	}

	mkNode(g, g.root, rs.vars, rs.unexportedVars, dryrun, true)
	if g.root.status == nodeStatusFailed {
		os.Exit(1)
	}
}

