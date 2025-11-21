package ml

import (
	"bufio"
	"io"
)

type Tokenizer interface {
	Next() (Token, error)
}

func NewTokenizer(reader io.Reader) Tokenizer {
	return &tokenizer{reader: *bufio.NewReader(reader)}
}

type Token struct {
	Name  string
	Type  int
	Value string
	Start int
	End   int
}

const (
	Block = iota
	End
	Text
)

type tokenizer struct {
	reader   bufio.Reader
	position int
}

func (t *tokenizer) Next() (Token, error) {
	startSymbol, err := t.reader.ReadByte()
	if err != nil {
		return Token{}, err
	}
	t.position++

	for {
		if t.isWhitespace(startSymbol) {
			startSymbol, err = t.reader.ReadByte()
			if err != nil {
				return Token{}, err
			}
			t.position++
			continue
		}

		switch startSymbol {
		case '/':
			blockToken := Token{Start: t.position, Name: "block", Type: Block}

			for {
				symbol, err := t.reader.ReadByte()
				if err != nil {
					return blockToken, err
				}
				t.position++

				if symbol == '/' {
					blockToken.End = t.position
					break
				}

				blockToken.Value += string(symbol)
			}

			return blockToken, nil
		default:
			textToken := Token{Start: t.position, Name: "text", Type: Text, Value: string(startSymbol)}

			for {
				symbol, err := t.reader.ReadByte()
				if err != nil {
					return textToken, err
				}
				t.position++

				if symbol == '/' {
					textToken.End = t.position
					break
				}

				textToken.Value += string(symbol)
			}

			return textToken, nil
		}
	}
}

func (t *tokenizer) isWhitespace(symbol byte) bool {
	switch symbol {
	case ' ', '\t', '\n', '\r':
		return true
	default:
		return false
	}
}
