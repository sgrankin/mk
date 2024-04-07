// String substitution and expansion.

package main

import (
	"os"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Expand a word. This includes substituting variables and handling quotes.
func expand(input string, vars map[string][]string, expandBackticks bool) []string {
	parts := make([]string, 0)
	expanded := ""
	var i, j int
	for i = 0; i < len(input); {
		j = strings.IndexAny(input[i:], "\"'`$\\")

		if j < 0 {
			expanded += input[i:]
			break
		}
		j += i

		expanded += input[i:j]
		c, w := utf8.DecodeRuneInString(input[j:])
		i = j + w

		var off int
		var out string
		switch c {
		case '\\':
			out, off = expandEscape(input[i:])
			expanded += out

		case '"':
			out, off = expandDoubleQuoted(input[i:], vars, expandBackticks)
			expanded += out

		case '\'':
			out, off = expandSingleQuoted(input[i:])
			expanded += out

		case '`':
			if expandBackticks {
				var outparts []string
				outparts, off = expandBackQuoted(input[i:], vars)
				if len(outparts) > 0 {
					outparts[0] = expanded + outparts[0]
					expanded = outparts[len(outparts)-1]
					parts = append(parts, outparts[:len(outparts)-1]...)
				}
			} else {
				out = input
				off = len(input)
				expanded += out
			}

		case '$':
			var outparts []string
			outparts, off = expandSigil(input[i:], vars)
			if len(outparts) > 0 {
				firstpart := expanded + outparts[0]
				if len(outparts) > 1 {
					parts = append(parts, firstpart)
					if len(outparts) > 2 {
						parts = append(parts, outparts[1:len(outparts)-1]...)
					}
					expanded = outparts[len(outparts)-1]
				} else {
					expanded = firstpart
				}
			}
		}

		i += off
	}

	if len(expanded) > 0 {
		parts = append(parts, expanded)
	}

	return parts
}

// Expand following a '\\'
func expandEscape(input string) (string, int) {
	c, w := utf8.DecodeRuneInString(input)
	if c == '\t' || c == ' ' {
		return string(c), w
	}
	if c == '\n' {
		return "", w
	}
	return "\\" + string(c), w
}

// Expand a double quoted string starting after a '\"'
func expandDoubleQuoted(input string, vars map[string][]string, expandBackticks bool) (string, int) {
	// find the first non-escaped "
	i := 0
	j := 0
	for {
		j = strings.IndexAny(input[i:], "\"\\")
		if j < 0 {
			break
		}
		j += i

		c, w := utf8.DecodeRuneInString(input[j:])
		i = j + w

		if c == '"' {
			return strings.Join(expand(input[:j], vars, expandBackticks), " "), i
		}

		if c == '\\' {
			if i < len(input) {
				_, w := utf8.DecodeRuneInString(input[i:])
				i += w
			} else {
				break
			}
		}
	}

	return input, len(input)
}

// Expand a single quoted string starting after a '\”
func expandSingleQuoted(input string) (string, int) {
	j := strings.Index(input, "'")
	if j < 0 {
		return input, len(input)
	}
	return input[:j], (j + 1)
}

var expandSigil_namelist_pattern = regexp.MustCompile(`^\s*([^:]+)\s*:\s*([^%]*)%([^=]*)\s*=\s*([^%]*)%([^%]*)\s*`)

// Expand something starting with at '$'.
func expandSigil(input string, vars map[string][]string) ([]string, int) {
	c, w := utf8.DecodeRuneInString(input)
	var offset int
	var varname string
	namelist_pattern := expandSigil_namelist_pattern

	if c == '$' { // escaping of "$" with "$$"
		return []string{"$"}, 2
	} else if c == '{' { // match bracketed expansions: ${foo}, or ${foo:a%b=c%d}
		j := strings.IndexRune(input[w:], '}')
		if j < 0 {
			return []string{"$" + input}, len(input)
		}
		varname = input[w : w+j]
		offset = w + j + 1

		// is this a namelist?
		mat := namelist_pattern.FindStringSubmatch(varname)
		if mat != nil && isValidVarName(mat[1]) {
			// ${varname:a%b=c%d}
			varname = mat[1]
			a, b, c, d := mat[2], mat[3], mat[4], mat[5]
			values, ok := vars[varname]
			if !ok {
				return []string{}, offset
			}

			pat := regexp.MustCompile(strings.Join([]string{`^\Q`, a, `\E(.*)\Q`, b, `\E$`}, ""))
			expanded_values := make([]string, 0, len(values))
			for _, value := range values {
				value_match := pat.FindStringSubmatch(value)
				if value_match != nil {
					expanded_values = append(expanded_values, expand(strings.Join([]string{c, value_match[1], d}, ""), vars, false)...)
				} else {
					// What case is this?
					expanded_values = append(expanded_values, value)
				}
			}

			return expanded_values, offset
		}
	} else { // bare variables: $foo
		// try to match a variable name
		i := 0
		j := i
		for j < len(input) {
			c, w = utf8.DecodeRuneInString(input[j:])
			if !(isalpha(c) || c == '_' || (j > i && isdigit(c))) {
				break
			}
			j += w
		}

		if j > i {
			varname = input[i:j]
			offset = j
		} else {
			offset = j + 1
			return []string{"$" + input[:offset]}, offset
			// return []string{"$" + input}, len(input)
		}
	}

	if isValidVarName(varname) {
		varvals, ok := vars[varname]
		if ok {
			return varvals, offset
		}

		// Find the subsitution in the environment.
		if varval, ok := os.LookupEnv(varname); ok {
			return []string{varval}, offset
		}

		return []string{"$" + input[:offset]}, offset
	}

	// Find the subsitution in the environment.
	if varval, ok := os.LookupEnv(varname); ok {
		return []string{varval}, offset
	}

	return []string{"$" + input}, len(input)
}

// Find and expand all sigils.
func expandSigils(input string, vars map[string][]string) []string {
	parts := make([]string, 0)
	expanded := ""
	for i := 0; i < len(input); {
		j := strings.IndexRune(input[i:], '$')
		if j < 0 {
			expanded += input[i:]
			break
		}
		j += i

		ex, k := expandSigil(input[j+1:], vars)
		if len(ex) > 0 {
			ex[0] = expanded + ex[0]
			expanded = ex[len(ex)-1]
			parts = append(parts, ex[:len(ex)-1]...)
		}
		i = k
	}

	if len(expanded) > 0 {
		parts = append(parts, expanded)
	}

	return parts
}

// Find and expand all sigils in a recipe, producing a flat string.
func expandRecipeSigils(input string, vars map[string][]string) string {
	expanded := ""
	for i := 0; i < len(input); {
		off := strings.IndexAny(input[i:], "$\\")
		if off < 0 {
			expanded += input[i:]
			break
		}
		expanded += input[i : i+off]
		i += off

		c, w := utf8.DecodeRuneInString(input[i:])
		if c == '$' {
			i += w
			ex, k := expandSigil(input[i:], vars)
			expanded += strings.Join(ex, " ")
			i += k
		} else if c == '\\' {
			i += w
			c, w := utf8.DecodeRuneInString(input[i:])
			if c == '$' {
				expanded += "$"
			} else {
				expanded += "\\" + string(c)
			}
			i += w
		}
	}

	return expanded
}

// Expand all unescaped '%' characters.
func expandSuffixes(input string, stem string) string {
	expanded := make([]byte, 0)
	for i := 0; i < len(input); {
		j := strings.IndexAny(input[i:], "\\%")
		if j < 0 {
			expanded = append(expanded, input[i:]...)
			break
		}

		c, w := utf8.DecodeRuneInString(input[j:])
		expanded = append(expanded, input[i:j]...)
		if c == '%' {
			expanded = append(expanded, stem...)
			i = j + w
		} else {
			j += w
			c, w := utf8.DecodeRuneInString(input[j:])
			if c == '%' {
				expanded = append(expanded, '%')
				i = j + w
			}
		}
	}

	return string(expanded)
}

// Expand a backtick quoted string, by executing the contents.
func expandBackQuoted(input string, vars map[string][]string) ([]string, int) {
	// TODO: expand sigils?
	j := strings.Index(input, "`")
	if j < 0 {
		return []string{input}, len(input)
	}

	env := os.Environ()
	for key, values := range vars {
		env = append(env, key+"="+strings.Join(values, " "))
	}

	// TODO - might have $shell available by now, but maybe not?
	// It's not populated, regardless

	var shell string
	var shellargs []string
	if len(vars["shell"]) < 1 {
		shell, shellargs = expandShell(defaultShell, shellargs)
	} else {
		shell, shellargs = expandShell(vars["shell"][0], shellargs)
	}

	// TODO: handle errors
	output, _ := subprocess(shell, shellargs, env, input[:j], true)

	parts := make([]string, 0)
	_, tokens := lexWords(output)
	for t := range tokens {
		parts = append(parts, t.val)
	}

	return parts, (j + 1)
}

// Expand the shell command into cmd, args...
// Ex. "sh -c", "pwd" becomes sh, [-c, pwd]
func expandShell(shcmd string, args []string) (string, []string) {
	var shell string
	var shellargs []string

	fields := strings.Fields(shcmd)
	shell = fields[0]

	if len(fields) > 1 {
		shellargs = fields[1:]
	}

	switch {
	// TODO - This case logic might be shaky, works for now
	case len(shellargs) > 0 && len(args) > 0:
		args = append(shellargs, args...)

	case len(shellargs) > 0 && dontDropArgs:
		args = append(shellargs, args...)

	default:
		// fmt.Println("dropping in expand!")
	}

	if len(shellargs) > 0 && dontDropArgs {
	} else {
	}

	return shell, args
}
