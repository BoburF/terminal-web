package main

import (
	"errors"
	"log"
	"os"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/net/html"
)

const (
	RootPath = "./resume/"
)

func main() {
	file, err := os.OpenFile(RootPath+"index.html", os.O_RDONLY, 0o644)
	if err != nil {
		log.Fatalln(err)
	}

	doc, err := html.Parse(file)
	if err != nil {
		log.Fatalln(err)
	}

	for node := range doc.Descendants() {
		if node.Data == "head" {
			foundScriptToBind(node)
		}

		if node.Data == "body" {
			// Parse all elements from the body
			_, _, inputs, buttons, err := drawTui(node)
			if err != nil {
				log.Fatalln(err)
			}

			// Create initial state
			initialState := State{
				boxes:   make([]Box, 0),
				inputs:  inputs,
				buttons: buttons,
				focused: 0,
			}

			p := tea.NewProgram(initialState)
			if _, err := p.Run(); err != nil {
				log.Fatalln(err)
			}

			return
		}
	}
}

func foundHTMLNode(doc *html.Node, nodeName string, attrName string) (*html.Node, error) {
	for node := range doc.ChildNodes() {
		if node.Data == nodeName && slices.ContainsFunc(node.Attr, func(attr html.Attribute) bool {
			return attr.Key == attrName
		}) {
			return node, nil
		}
	}

	return nil, errors.New("didn't found the tag")
}

func foundHTMLNodeWithAttr(doc *html.Node, nodeName string, attrName string, attrValue string) (*html.Node, error) {
	for node := range doc.ChildNodes() {
		if node.Data == nodeName && slices.ContainsFunc(node.Attr, func(attr html.Attribute) bool {
			return attr.Key == attrName && attr.Val == attrValue
		}) {
			return node, nil
		}
	}

	return nil, errors.New("didn't found the tag")
}

func foundAttr(attr []html.Attribute, attrName string) (html.Attribute, error) {
	for _, attr := range attr {
		if attr.Key == attrName {
			return attr, nil
		}
	}

	return html.Attribute{}, errors.New("didn't found the attr")
}
