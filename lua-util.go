package main

import "github.com/Shopify/go-lua"

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
