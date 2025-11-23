package ml

import (
	"bufio"
	"fmt"
	"io"
)

const (
	Block    = "block"
	EndBlock = "end"
	Text     = "text"
)

type Tokenizer interface {
	Next() (Token, error)
	IsNextAvailable() bool
}

func NewTokenizer(reader io.Reader) Tokenizer {
	return &tokenizer{reader: *bufio.NewReader(reader)}
}

type Token struct {
	Name  string
	Value string
	Start int
	End   int
}

type tokenizer struct {
	reader   bufio.Reader
	position int
}

func (t *tokenizer) IsNextAvailable() bool {
	_, err := t.reader.Peek(1)
	return err == nil
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
			blockToken := Token{Start: t.position, Name: Block}

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

			blockToken.Name = blockToken.Value

			return blockToken, nil
		default:
			textToken := Token{Start: t.position, Name: Text, Value: string(startSymbol)}

			for {
				symbol, err := t.reader.Peek(1)
				if err != nil {
					return textToken, err
				}

				if symbol[0] == '/' {
					textToken.End = t.position
					break
				}

				sym, err := t.reader.ReadByte()
				if err != nil {
					return textToken, err
				}
				t.position++

				textToken.Value += string(sym)
			}
			fmt.Println("value", textToken.Value)

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
