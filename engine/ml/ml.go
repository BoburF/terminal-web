// Package ml (Markup Language) tokenizer, lexer, parser, combined (compilor?)
// It will be used for parsing the content to AST
package ml

type ml struct {
	Tokenizer Tokenizer
	Lexer     lexer
}

func NewMlParser() ml {
	return ml{}
}
