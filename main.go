package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/net/html"
	"golang.org/x/term"
)

const (
	RootPath = "./resume/"
)

func main() {
	serverMode := flag.Bool("server", false, "Run as SSH server")
	port := flag.String("port", "", "SSH server port (overrides default 4569)")
	flag.Parse()

	if *serverMode {
		// Load default configuration
		config := DefaultConfig()

		// Override port if provided
		if *port != "" {
			config.Server.Port = *port
		}

		sshServer, err := NewSSHServer(config)
		if err != nil {
			log.Fatalf("Failed to initialize SSH server: %v", err)
		}

		if err := sshServer.Start(); err != nil {
			log.Fatalf("Failed to start SSH server: %v", err)
		}
		return
	}

	runLocalMode()
}

func runLocalMode() {
	fd := int(os.Stdout.Fd())

	if !term.IsTerminal(fd) {
		fmt.Println("Not a terminal, cannot get size.")
		return
	}

	width, height, err := term.GetSize(fd)
	if err != nil {
		fmt.Printf("Error getting terminal size: %v\n", err)
		return
	}

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
			state, err := drawTui(node)
			if err != nil {
				log.Fatalln(err)
			}
			state.Width = width
			state.Height = height

			p := tea.NewProgram(state)
			if _, err := p.Run(); err != nil {
				log.Fatalln(err)
				os.Exit(1)
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

func foundAttr(attr *[]html.Attribute, attrName string) (html.Attribute, error) {
	for _, attr := range *attr {
		if attr.Key == attrName {
			return attr, nil
		}
	}

	return html.Attribute{}, errors.New("didn't found the attr")
}
