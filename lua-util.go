package main

import (
	"io"
	"log"
	"os"

	"github.com/Shopify/go-lua"
	"golang.org/x/net/html"
)

const (
	Link = "link"
	Href = "href"
)

func foundScriptToBind(node *html.Node) {
	sciptBinding, err := foundHTMLNode(node, Link, Href)
	if err != nil {
		log.Fatalln(err)
	}

	pathAttr, err := foundAttr(sciptBinding.Attr, Href)
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

	err = luaRegister(string(script))
	if err != nil {
		log.Fatalln(err)
	}
}

func luaRegister(luaScript string) error {
	l := lua.NewState()
	lua.OpenLibraries(l)

	l.Register("bind", func(state *lua.State) int {
		return 0
	})

	l.Register("quit", func(state *lua.State) int {
		return 0
	})

	return lua.DoString(l, luaScript)
}
