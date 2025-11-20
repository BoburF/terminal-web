// Package ml (Markup Language) tokenizer, lexer, parser, combined (compilor?)
// It will be used for parsing the content to AST
package ml

import "io"

type ml struct {
	Tokenizer Tokenizer
	Lexer     lexer
}

func NewMlParser(reader io.Reader) ml {
	return ml{Tokenizer: NewTokenizer(reader)}
}
