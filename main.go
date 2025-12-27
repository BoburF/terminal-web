package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"slices"

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
			sciptBinding, err := foundHTMLNode(node, "link", "href")
			if err != nil {
				log.Fatalln(err)
			}

			pathAttr, err := foundAttr(sciptBinding.Attr, "href")
			if err != nil {
				log.Fatalln(err)
			}

			file, err := os.OpenFile(RootPath+pathAttr.Val, os.O_RDONLY, 0o644)
			if err != nil {
				log.Fatalln(err)
			}

			script, err := io.ReadAll(file)
			if err != nil {
				log.Fatalln(err)
			}

			fmt.Println(pathAttr)

			err = luaRegister(string(script))
			if err != nil {
				log.Fatalln(err)
			}
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

func foundAttr(attr []html.Attribute, attrName string) (html.Attribute, error) {
	for _, attr := range attr {
		if attr.Key == attrName {
			return attr, nil
		}
	}

	return html.Attribute{}, errors.New("didn't found the attr")
}
