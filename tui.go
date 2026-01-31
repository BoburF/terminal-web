package main

import (
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"golang.org/x/net/html"
)

func drawTui(doc *html.Node) (State, error) {
	boxes := parseMain(doc)

	controllers := parseControlls(doc)

	return State{boxes: boxes, interactivity: controllers}, nil
}

func parseControlls(doc *html.Node) []Controller {
	controllers := make([]Controller, 0)

	controllersSection, err := foundHTMLNodeWithAttr(doc, "div", "class", "controllers")
	if err != nil {
		log.Fatalln(err)
	}

	for node := range controllersSection.ChildNodes() {
		if strings.TrimSpace(node.Data) == "" {
			continue
		}

		button := Controller{}

		switch node.Data {
		case "button":
			text := getText(node.FirstChild)
			controllType, err := foundAttr(&node.Attr, "type")
			if err != nil {
				log.Fatalln(err)
			}
			bindingType, err := foundAttr(&node.Attr, "bind")
			if err != nil {
				log.Fatalln(err)
			}

			button.name = text
			button.event = controllType.Val
			button.combination = bindingType.Val

			controllers = append(controllers, button)
		}
	}

	return controllers
}

func parseMain(doc *html.Node) []Box {
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
			if strings.TrimSpace(childNode.Data) == "" {
				continue
			}

			switch childNode.Type {
			case html.ElementNode:
				switch childNode.Data {
				case "input":
					input := textinput.New()

					placeholderValue, err := foundAttr(&childNode.Attr, "value")
					if err != nil {
						log.Fatalln(err)
					}

					input.Placeholder = placeholderValue.Val

					box.inputs = append(box.inputs, input)
					box.context = append(box.context, input)
				case "h1":
					text := getText(childNode.FirstChild)
					box.isNotEmplty = true
					box.texts = append(box.texts, text)
					box.context = append(box.context, text)
				case "p":
					text := getText(childNode.FirstChild)
					box.isNotEmplty = true
					box.texts = append(box.texts, text)
					box.context = append(box.context, text)
				}
			default:
				continue
			}
		}

		boxes = append(boxes, box)
	}

	return boxes
}

func getText(node *html.Node) string {
	if strings.TrimSpace(node.Data) == "" {
		return getText(node.NextSibling)
	}

	return strings.TrimSpace(node.Data)
}
