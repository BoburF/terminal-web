package main

import (
	"io"
	"log"
	"os"
	"sync"

	"github.com/Shopify/go-lua"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/net/html"
)

const (
	Link = "link"
	Href = "href"
)

var (
	luaBindingRegistry = make(map[string]func() tea.Cmd)
	luaRegistryMutex   sync.RWMutex
	luaState           *lua.State
)

func foundScriptToBind(node *html.Node) {
	sciptBinding, err := foundHTMLNode(node, Link, Href)
	if err != nil {
		log.Fatalln(err)
	}

	pathAttr, err := foundAttr(&sciptBinding.Attr, Href)
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
	luaState = lua.NewState()
	lua.OpenLibraries(luaState)

	luaState.Register("bind", func(state *lua.State) int {
		if state.Top() < 2 {
			return 0
		}

		key := lua.CheckString(state, 1)

		luaRegistryMutex.Lock()
		if key == "ctrl+c" || key == "q" {
			luaBindingRegistry[key] = func() tea.Cmd {
				return tea.Quit
			}
		} else {
			luaBindingRegistry[key] = func() tea.Cmd {
				return nil
			}
		}
		luaRegistryMutex.Unlock()

		return 0
	})

	luaState.Register("quit", func(state *lua.State) int {
		return 0
	})

	return lua.DoString(luaState, luaScript)
}

func GetLuaBinding(key string) (func() tea.Cmd, bool) {
	luaRegistryMutex.RLock()
	defer luaRegistryMutex.RUnlock()
	binding, ok := luaBindingRegistry[key]
	return binding, ok
}

func ClearLuaBindings() {
	luaRegistryMutex.Lock()
	defer luaRegistryMutex.Unlock()
	luaBindingRegistry = make(map[string]func() tea.Cmd)
}
