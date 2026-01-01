package main

import (
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"golang.org/x/net/html"
)

func drawTui(doc *html.Node) (State, error) {
	boxes := make([]Box, 0)

	main, err := foundHTMLNodeWithAttr(doc, "div", "class", "main")
	if err != nil {
		log.Fatalln(err)
	}

	for node := range main.ChildNodes() {
		if strings.TrimSpace(node.Data) == "" {
			continue
		}

		box := Box{
			isNotEmplty: false,
			texts:       make([]string, 0),
			inputs:      make([]textinput.Model, 0),
			context:     make([]any, 0),
		}

		for childNode := range node.ChildNodes() {
			if strings.TrimSpace(node.Data) == "" {
				continue
			}

			switch childNode.Type {
			case html.TextNode:
				box.isNotEmplty = true
				text := getText(childNode)
				box.texts = append(box.texts, text)
				box.context = append(box.context, text)
			case html.ElementNode:
				if childNode.Data == "input" {
					input := textinput.New()

					placeholderValue, err := foundAttr(&childNode.Attr, "value")
					if err != nil {
						log.Fatalln(err)
					}

					input.Placeholder = placeholderValue.Val

					box.inputs = append(box.inputs, input)
					box.context = append(box.context, input)
				}
			default:
				continue
			}
		}

		boxes = append(boxes, box)
	}

	return State{boxes: boxes}, nil
}

func getText(node *html.Node) string {
	if strings.TrimSpace(node.Data) == "" {
		return getText(node.NextSibling)
	}

	return node.Data
}
