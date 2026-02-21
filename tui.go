package main

import (
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"golang.org/x/net/html"

	"github.com/BoburF/terminal-web.git/internal/page"
)

func drawTui(doc *html.Node) (State, error) {
	boxes, sectionTitles, pageLinks := parseMain(doc)

	controllers := parseControlls(doc)

	return State{boxes: boxes, interactivity: controllers, sectionTitles: sectionTitles, pageLinks: pageLinks}, nil
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
			text = strings.Join(strings.Fields(text), " ")
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

func parseMain(doc *html.Node) ([]Box, []string, []page.PageLink) {
	boxes := make([]Box, 0)
	sectionTitles := make([]string, 0)
	pageLinks := make([]page.PageLink, 0)

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

		pageLink := page.PageLink{}

		sectionType := ""
		if typeAttr, err := foundAttr(&node.Attr, "section-type"); err == nil {
			sectionType = typeAttr.Val
		}

		if sectionType == "page-link" {
			if targetAttr, err := foundAttr(&node.Attr, "page-target"); err == nil {
				pageLink.Target = targetAttr.Val
				box.IsPageLink = true
				box.PageTarget = targetAttr.Val
			}
		}

		sectionTitle := "Section"
		if titleAttr, err := foundAttr(&node.Attr, "section-title"); err == nil {
			sectionTitle = titleAttr.Val
			pageLink.SectionTitle = titleAttr.Val
		}
		sectionTitles = append(sectionTitles, sectionTitle)
		pageLinks = append(pageLinks, pageLink)

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
					if childNode.FirstChild != nil {
						text := getText(childNode.FirstChild)
						text = strings.Join(strings.Fields(text), " ")
						box.isNotEmplty = true
						box.texts = append(box.texts, text)
						box.context = append(box.context, text)
					}
				case "p":
					if childNode.FirstChild != nil {
						text := getText(childNode.FirstChild)
						text = strings.Join(strings.Fields(text), " ")
						box.isNotEmplty = true
						box.texts = append(box.texts, text)
						box.context = append(box.context, text)
					}
				}
			default:
				continue
			}
		}

		boxes = append(boxes, box)
	}

	return boxes, sectionTitles, pageLinks
}

func getText(node *html.Node) string {
	if node == nil {
		return ""
	}
	if strings.TrimSpace(node.Data) == "" {
		return getText(node.NextSibling)
	}

	return node.Data
}
