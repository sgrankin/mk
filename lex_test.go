package main

import "testing"

func TestTokenTypeString(t *testing.T) {
	tests := []struct {
		name string
		typ  tokenType
		want string
	}{
		{"error", tokenError, "[Error]"},
		{"newline", tokenNewline, "[Newline]"},
		{"word", tokenWord, "[Word]"},
		{"pipe_include", tokenPipeInclude, "[PipeInclude]"},
		{"redir_include", tokenRedirInclude, "[RedirInclude]"},
		{"colon", tokenColon, "[Colon]"},
		{"assign", tokenAssign, "[Assign]"},
		{"recipe", tokenRecipe, "[Recipe]"},
		{"assign_u", tokenAssignU, "[AssignU]"},
		{"unknown", tokenType(999), "[MysteryToken]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.typ.String(); got != tt.want {
				t.Errorf("tokenType(%d).String() = %q, want %q", tt.typ, got, tt.want)
			}
		})
	}
}

func TestTokenString(t *testing.T) {
	// Cover the error and newline special cases in token.String().
	errTok := token{typ: tokenError, val: "some error"}
	if got := errTok.String(); got != "some error" {
		t.Errorf("error token String() = %q, want %q", got, "some error")
	}
	nlTok := token{typ: tokenNewline}
	if got := nlTok.String(); got != "\\n" {
		t.Errorf("newline token String() = %q, want %q", got, "\\n")
	}
}

func TestLexerNextEOF(t *testing.T) {
	// Cover next() returning eof when pos >= len(input).
	l := &lexer{input: ""}
	if got := l.next(); got != eof {
		t.Errorf("next() on empty input = %q, want eof", got)
	}
}

func TestSkipUntilEOF(t *testing.T) {
	// Cover skipUntil hitting EOF.
	// Construct lexer directly (bypassing lex() which adds trailing newline).
	output := make(chan token, 1)
	l := &lexer{input: "no newline here", line: 1, output: output}
	l.skipUntil("\n")
	tok := <-output
	if tok.typ != tokenError {
		t.Errorf("skipUntil EOF: got token type %v, want tokenError", tok.typ)
	}
}
