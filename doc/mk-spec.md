# mk Specification

This document describes the specification of Plan 9 mk, derived from:
- The mk(1) man page (plan9port)
- "Maintaining Files on Plan 9 with Mk" (Hume & Flandrena)
- The Plan 9 from User Space C source code

Where our Go implementation diverges from Plan 9, differences are noted inline as **[DIVERGENCE]**.

## 1. Overview

Mk reads a file (mkfile) describing dependencies between files and executes
commands to bring them up to date. A mkfile contains three types of statements:
**assignments**, **includes**, and **rules**.

## 2. Lexical Structure

### 2.1 Lines

Mk processes input line by line. A logical line may span multiple physical lines
via backslash-newline continuation: the sequence `\<newline>` is deleted
(in non-recipe context) or replaced with a space (in recipe context).

Blank lines are ignored. Lines beginning with `#` (outside quotes) are comments
and are discarded up to the next newline.

### 2.2 Character Classes

Plan 9 mk defines word characters via `WORDCHR` (mk.h):

```c
#define WORDCHR(r) ((r) > ' ' && !utfrune("!\"#$%&'()*+,-./:;<=>?@[\\]^`{|}~", (r)))
```

Characters matching `WORDCHR` are: alphanumerics, underscore, and any character
with code point > 127 (Unicode). All ASCII punctuation terminates a word in
variable-name contexts.

**[DIVERGENCE]** Our `nonBareRunes` is much smaller: `` " \t\n\r\\=:#'\"$` ``.

### 2.3 Quoting

Three quoting mechanisms:

| Quote | Behavior |
|-------|----------|
| `'...'` | Literal string. No escapes. Terminated by matching `'`. |
| `"..."` | Interpreted string. `\"` escapes a literal quote, `\\` escapes backslash. Variables and backticks are expanded inside. |
| `` `...` `` | Command substitution (sh-style). See §2.5. |
| `` `{...} `` | Command substitution (rc-style). See §2.5. |

Single and double quotes can appear adjacent to bare words and to each other,
forming a single token: `foo'bar'"baz"` → `foobarbaz`.

### 2.4 Comments

A `#` character outside quotes begins a comment. Everything from `#` to the
end of the line is discarded.

If the last non-whitespace character before the newline is `\`,
the comment-ending newline is treated as a continuation (the next line is also
discarded).

### 2.5 Backtick (Command Substitution)

Backtick substitution executes a shell command and substitutes its output.

In Plan 9 mk (lex.c `bquote()`), backtick expansion happens **at lex time**
during line assembly (`assline()`). The lexer encounters `` ` ``, then:

1. Skips whitespace after the opening backtick.
2. If the next character is `{`, uses `}` as the terminator (rc-style: `` `{cmd} ``).
3. Otherwise, uses `` ` `` as the terminator (sh-style: `` `cmd` ``).
4. Reads until the terminator, handling nested quotes (`'`, `"`, `\`).
5. Executes the command via `execsh()`.
6. The command's output replaces the backtick expression in the line buffer.

The output is then subject to normal word splitting.

Backtick can appear in the middle of a bare word: `foo`echo bar`` produces
`foobar` (the output of `echo bar` is `bar`, concatenated with `foo`).

**[DIVERGENCE]** Expansion happens in `expand.go`, not at lex time. This is
architecturally different but functionally equivalent for most cases.

### 2.6 Variable References

Two forms:
- `$name` — bare variable reference; `name` is the longest sequence of `WORDCHR` characters.
- `${name}` — bracketed variable reference.
- `${name:A%B=C%D}` — namelist pattern substitution (see §4).

`$$` produces a literal `$`.

In recipes, `\$` also produces a literal `$` (shell escaping).

### 2.7 Line Classification

The first unquoted occurrence of `:`, `=`, or `<` determines the line type:

| Character | Line type |
|-----------|-----------|
| `:` | Rule |
| `=` | Assignment |
| `<` | Include (file) |
| `<\|` | Include (pipe/command) |

If none of these appears, it's a syntax error.

## 3. Assignments

```
name = value
name =U= value
```

The left side must be a single word (the variable name). The right side is
split into words which become the variable's value (a list).

The `U` attribute between `=` signs marks the variable as unexported
(not placed in the environment for recipes).

### 3.1 Precedence

Variable values have this precedence (highest first):
1. Command-line assignment (`mk name=value`)
2. Assignment statement in mkfile
3. Inherited from environment
4. Implicitly set by mk

### 3.2 Evaluation Timing

- Assignment values are evaluated **when read** (eagerly).
- Recipe bodies are evaluated **when executed** (lazily, by the shell).
- This means the last assignment wins for recipes, but earlier assignments
  affect subsequent rule headers.

## 4. Namelist Substitution

```
${var:A%B=C%D}
```

For each word in `$var` matching the pattern `A%B` (where `%` matches any
string), substitute `C%D` (replacing `%` with the matched portion).
`A`, `B`, `C`, `D` may be empty.

Equivalent to applying `s/A(.*)B/C\1D/` to each word.

Example: `${SRC:%.c=%.o}` transforms `a.c b.c` → `a.o b.o`.

## 5. Include Statements

### 5.1 File Include

```
<filename
```

Replaces itself with the contents of `filename`. Variables in `filename` are
expanded before opening. The included file is parsed recursively.

### 5.2 Pipe Include

```
<|command args
```

Executes `command args` via the shell, then parses the output as mkfile content.

## 6. Rules

### 6.1 Syntax

```
targets : attributes : prerequisites
    recipe
```

- **targets**: one or more words (file names or patterns).
- **attributes**: zero or more attribute characters between the colons.
- **prerequisites**: zero or more words.
- **recipe**: zero or more lines, each beginning with whitespace (tab or space).

The second colon is required if attributes are present. If no attributes,
the form is `targets : prerequisites`.

### 6.2 Recipe Execution

The entire recipe is passed to the shell as a single script (not line-by-line
like make). The first whitespace character on each line is stripped.

By default, the shell is invoked with `-e` (exit on error). The `E` attribute
overrides this.

### 6.3 Automatic Variables

Available in recipes (set by mk before execution):

| Variable | Value |
|----------|-------|
| `$target` | The target being built |
| `$prereq` | All prerequisites |
| `$newprereq` | Out-of-date prerequisites only |
| `$stem` | String matched by `%` or `&` in metarule |
| `$stemN` | Nth regex subexpression match (R attribute) |
| `$alltarget` | All targets of the rule |
| `$newmember` | Archive member names from `$newprereq` |
| `$nproc` | Slot number (0-based) of this parallel job |
| `$pid` | Process ID of mk |

### 6.4 Attributes

| Attr | Name | Effect |
|------|------|--------|
| `D` | Delete | Remove target file if recipe fails |
| `E` | Errors | Don't pass `-e` to shell; continue on errors |
| `N` | No-recipe | Target need not exist and has no recipe |
| `n` | No-virtual | Metarule matches only non-virtual prereqs |
| `P` | Program | Custom out-of-date test: `Pprog:` runs `prog target prereq` |
| `Q` | Quiet | Don't print recipe before execution |
| `R` | Regexp | Target is a regular expression |
| `U` | Update | Force target timestamp update even if not changed |
| `V` | Virtual | Target is not a file; always considered out of date |

**[DIVERGENCE]** Our implementation adds:
- `S` — Shell: specify alternative shell for this rule (e.g., `S"awk"`)
- `X` — Exclusive: recipe acquires all parallel job slots before executing

### 6.5 No-Recipe Rules

A rule with prerequisites but no recipe adds those prerequisites to all other
rules with the same target that DO have recipes. This is the mechanism for
adding header file dependencies:

```
%.o: hdr.h          # no recipe — adds hdr.h as prereq
%.o: %.c            # has recipe — hdr.h is added here too
    cc $stem.c
```

### 6.6 No-Prerequisite Rules

A rule with no prerequisites and a recipe:
- For a real target: executes only if the target file doesn't exist.
- For a virtual target: always executes.

A rule with no prerequisites and no recipe is a "no-effect" rule — it is
ignored during graph construction.

## 7. Metarules

### 7.1 Pattern Metarules

Two intrinsic patterns:

| Pattern | Matches | Equivalent regex |
|---------|---------|-----------------|
| `%` | One or more of anything | `.+` |
| `&` | One or more of anything except `/` and `.` | `[^./]+` |

Only one `%` or `&` may appear in a target pattern.

In prerequisites, `%` or `&` is replaced by the stem (the string matched
in the target).

### 7.2 Regex Metarules

With the `R` attribute, the target is interpreted as a regular expression.
Prerequisites can reference subexpressions with `\1` through `\9`.
Recipe variables `$stem1` through `$stem9` provide the same matches.

### 7.3 Metarule Evaluation

Each metarule is applied at most once per target (controlled by `NREP`,
default 1) to prevent infinite cycles.

A metarule with the `n` (NOVIRT) attribute does not apply when the only
matching rule for the target has the `V` (virtual) attribute.

## 8. Dependency Graph

### 8.1 Construction (`applyrules`)

For each target, mk builds a dependency graph:

1. **Concrete rules first**: Look up rules whose target literally matches.
   Skip "no-effect" rules (no recipe AND no prerequisites).
   For each matching rule, create arcs to prerequisite nodes (recursively).

2. **Metarules second**: Iterate all metarules. For each:
   - Skip no-effect metarules.
   - If `n` attribute and target only has virtual arcs, skip.
   - Match target against pattern (% or regex).
   - Create arcs to prerequisite nodes (with stem substitution).

3. A rule counter (`cnt[]`) limits each rule to `NREP` applications per target.

### 8.2 Cycle Detection (`cyclechk`)

DFS traversal; if a node is revisited while its `CYCLE` flag is set, mk
reports an error and exits.

### 8.3 Vacuous Node Pruning (`vacuous`)

A node is **vacuous** if:
- It is not PROBABLE (no concrete rule matched it, and it has no existing file).
- ALL of its prerequisites are also vacuous.

Vacuous arcs from metarules are marked `TOGO` and deleted. However, if a rule
produced BOTH vacuous and non-vacuous arcs, the vacuous ones are retained
(they came from the same rule).

### 8.4 Ambiguity Resolution (`ambiguous`)

After pruning, if a node has arcs from multiple rules with different recipes:
- **Concrete beats meta**: If one recipe comes from a concrete rule and
  another from a metarule, the metarule's arc is deleted.
- **Otherwise**: The recipes are ambiguous. Mk prints a trace and exits.

The trace format shows the dependency chain:
```
mk: ambiguous recipes for target:
    target <-(file:line)- prereq1 <-(file:line)- prereq2
    target <-(file:line)- prereq3
```

### 8.5 Attribute Propagation (`attribute`)

After ambiguity resolution, attributes from rules are propagated to nodes:
- `V` → node gets `VIRTUAL` flag, time set to 0
- `N` → node gets `NORECIPE` flag
- `D` → node gets `DELETE` flag

## 9. Execution

### 9.1 Work Algorithm (`work` in mk.c)

For each node (depth-first):

1. If already `BEINGMADE`, skip (parallel build in progress).
2. If `MADE` but `PRETENDING`, and parent is out of date, unpretend.
3. If no prerequisites and no file exists: error (don't know how to make).
4. If no prerequisites and file exists: mark `MADE`.
5. Recursively process prerequisites. Track which are out of date.
6. If all prerequisites are ready and node is out of date:
   - Check if we can **pretend** (missing intermediate optimization).
   - If not, execute the recipe via `dorecipe`.

### 9.2 Missing Intermediates (Pretending)

If a target is up to date with respect to the transitive prerequisites
(skipping a missing intermediate), mk pretends the intermediate exists
rather than rebuilding it. This is useful for archives where object files
are deleted after archiving.

The `-i` flag forces rebuild of missing intermediates, disabling this
optimization.

### 9.3 Parallel Execution

`$NPROC` controls the maximum number of concurrent recipes (default:
number of available processors).

Recipes for independent prerequisites execute in parallel. The `$nproc`
variable tells each recipe its slot number.

### 9.4 Recipe Output

Before execution, mk prints the recipe (unless `Q` attribute or `-q` flag).
The recipe is expanded with variable values for display.

In Plan 9 mk, the `front()` function truncates long recipe displays to
5 fields with `...` in the middle.

## 10. Shell Interface

### 10.1 Shell Structure

Plan 9 mk supports pluggable shells via the `Shell` struct:

```c
typedef struct Shell {
    char *name;
    char *termchars;    // chars that terminate assignment attributes
    int   iws;          // inter-word separator in environment
    char *(*charin)();  // find unescaped chars in string
    char *(*expandquote)(); // extract escaped token
    int   (*escapetoken)(); // input escaped token
    char *(*copyq)();   // handle quoted strings
    int   (*matchname)(); // does name match
} Shell;
```

The default is `sh` (Bourne shell). Plan 9 also supports `rc`.

### 10.2 sh Shell Conventions

- `shcharin()`: Searches for unescaped characters, respecting `\`, `'`, `"` quoting.
- `shexpandquote()`: Handles `\x` (single char escape), `'...'`, `"..."`.
- `shescapetoken()`: Input processing for escaped tokens during lexing.
- `shcopyq()`: Copies quoted strings including backtick strings `` `...` ``.
- `termchars`: `"'= \t` — characters that terminate assignment attributes.

### 10.3 MKSHELL Variable

Setting `MKSHELL` changes the shell used for subsequent recipe execution.

**[DIVERGENCE]** Our implementation uses a `shell` variable and supports
per-rule shell override via the `S` attribute.

## 11. Environment

### 11.1 Import/Export

On startup, mk imports all environment variables as mk variables.
Before executing each recipe, mk exports all variables to the environment.

Variables marked with the `U` attribute on assignment are not exported.

### 11.2 Internal Variables

These are set by mk and available in recipes:
`target`, `stem`, `prereq`, `pid`, `nproc`, `newprereq`, `alltarget`,
`newmember`, `stem0`–`stem9`.

### 11.3 Inter-Word Separator

List variables are exported to the environment using a separator character.
For `sh`, the separator is space (` `). For `rc`, it is `\x01` (SOH).

**[DIVERGENCE]** Our implementation always uses space.

## 12. Command-Line Flags

| Flag | Effect |
|------|--------|
| `-a` | Force rebuild of all targets (ignore timestamps) |
| `-f file` | Use `file` instead of `mkfile` |
| `-i` | Force rebuild of missing intermediates |
| `-k` | Continue after errors (keep going) |
| `-n` | Dry run (print recipes without executing) |
| `-q` | Don't print recipes before execution |
| `-r` | Shallow rebuild (only direct prerequisites) |
| `-t` | Touch targets instead of executing recipes |
| `-w target` | Treat `target` as recently modified |

**[DIVERGENCE]** Our implementation adds:
- `-p N` — Set parallelism level (default: number of CPUs)
- `-l N` — Max times a specific rule can be applied (default: 1)
- `-C dir` — Change to `dir` before reading mkfile
- `-F` — Don't drop shell args when no further arguments specified
- `-I` — Interactive mode: prompt before executing rules
- `-dot` — Print dependency graph in Graphviz dot format and exit
- `-color` — Enable/disable color output (default: auto-detect TTY)
- `-shell cmd` — Default shell (default: `sh -e`)

- `-e` — Explain why targets are out of date (prints staleness decisions to stderr)

## Appendix A: Known Divergences Summary

| Area | Plan 9 | Our Go Implementation |
|------|--------|----------------------|
| Backtick timing | Expanded at lex time | Expanded at eval time (in `expand.go`) |
| Shell interface | Pluggable Shell struct (sh, rc) | Fixed sh with `S` attribute for per-rule override |
| Environment separator | Shell-dependent (space for sh, \x01 for rc) | Always space |
| Shell variable | `MKSHELL` changes shell globally | Uses `shell` variable |
| Archive members | `lib(member)` syntax with special handling | Not supported |
| Recipe display | `front()` truncates to 5 fields | No truncation |
| `-e` flag | Explain pretend/intermediate logic only | Explains all staleness decisions |
