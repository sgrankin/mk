package main

import "testing"

func TestTokenTypeString(t *testing.T) {
	// Cover all named cases and the fallthrough default in tokenType.String().
	cases := []struct {
		typ  tokenType
		want string
	}{
		{tokenError, "[Error]"},
		{tokenNewline, "[Newline]"},
		{tokenWord, "[Word]"},
		{tokenPipeInclude, "[PipeInclude]"},
		{tokenRedirInclude, "[RedirInclude]"},
		{tokenColon, "[Colon]"},
		{tokenAssign, "[Assign]"},
		{tokenRecipe, "[Recipe]"},
		{tokenAssignU, "[AssignU]"},
		{tokenType(999), "[MysteryToken]"},
	}
	for _, tc := range cases {
		if got := tc.typ.String(); got != tc.want {
			t.Errorf("tokenType(%d).String() = %q, want %q", tc.typ, got, tc.want)
		}
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
