package ml

import "io"

type Tokenizer interface {
	Next() (Token, error)
}

type Token struct {
	Name  string
	Type  int
	Value string
	Start int
	End   int
}

func NewTokenizer(reader *io.Reader) Tokenizer {
	return tokenizer{reader: reader}
}

const (
	Block = iota
	End
	Text
)

type tokenizer struct {
	reader *io.Reader
}

func (t tokenizer) Next() (Token, error) {
	return Token{}, nil
}
