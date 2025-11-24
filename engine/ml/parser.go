package ml

import (
	"io"
)

const (
	Ast       = "AST"
	BlockNode = "block"
	TextNode  = "text"
)

type Parser interface {
	Analyze(tokenizer Tokenizer) (Node, error)
}

type Node struct {
	Type     string `json:"type"`
	Value    string `json:"value"`
	Children []Node `json:"children"`
}

func NewParser() Parser {
	return &parser{}
}

type parser struct{}

func (p *parser) Analyze(tokenizer Tokenizer) (Node, error) {
	ast := Node{Type: Ast, Value: Ast, Children: make([]Node, 0)}

	stack := []*Node{&ast}
	for tokenizer.IsNextAvailable() {
		token, err := tokenizer.Next()
		if err != nil {
			if err == io.EOF {
				return ast, nil
			}
			return ast, err
		}

		current := stack[len(stack)-1]

		switch token.Name {
		case Block:
			if token.Value != Block {
				continue
			}

			block := Node{Type: BlockNode, Children: make([]Node, 0)}
			current.Children = append(current.Children, block)
			stack = append(stack, &current.Children[len(current.Children)-1])
		case Text:
			current.Value = token.Value
		case EndBlock:
			if token.Value != EndBlock {
				continue
			}

			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
			}
		}
	}

	return ast, nil
}
