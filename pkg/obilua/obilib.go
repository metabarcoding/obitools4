package obilua

import lua "github.com/yuin/gopher-lua"

func RegisterObilib(luaState *lua.LState) {
	RegisterObiSeq(luaState)
	RegisterObiTaxonomy(luaState)
}
