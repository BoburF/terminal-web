package ml

type Lexer interface{}

func NewLexer() Lexer {
	return lexer{}
}

type lexer struct{}
