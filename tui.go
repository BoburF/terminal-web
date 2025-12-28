package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/net/html"
)

var boxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	Padding(1)

func extractTextContent(node *html.Node) string {
	if node.Type == html.TextNode {
		return strings.TrimSpace(node.Data)
	}

	var result strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		result.WriteString(extractTextContent(child))
	}
	return strings.TrimSpace(result.String())
}

func drawTui(node *html.Node) ([]string, []string, []textinput.Model, []string, error) {
	var boxes []string
	var text []string
	var inputs []textinput.Model
	var buttons []string

	switch node.Data {
	case "h1":
		content := extractTextContent(node)
		if content != "" {
			text = append(text, boxStyle.Render(content))
		}
	case "input":
		ti := textinput.New()
		ti.Width = 30

		for _, attr := range node.Attr {
			switch attr.Key {
			case "value", "placeholder":
				ti.Placeholder = attr.Val
			}
		}
		inputs = append(inputs, ti)
	case "button":
		content := extractTextContent(node)
		if content != "" {
			buttonStyle := lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				Padding(1, 1)

			buttons = append(buttons, buttonStyle.Render(content))
		}
	case "div":
		boxes = append(boxes, text...)
		boxes = append(boxes, buttons...)
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		childBoxes, childTexts, childInputs, childButtons, err := drawTui(child)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		boxes = append(boxes, childBoxes...)
		text = append(text, childTexts...)
		inputs = append(inputs, childInputs...)
		buttons = append(buttons, childButtons...)
	}

	return boxes, text, inputs, buttons, nil
}
