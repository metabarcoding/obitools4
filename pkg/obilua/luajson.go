package obilua

import (
	"encoding/json"

	lua "github.com/yuin/gopher-lua"
)

// RegisterJSON registers the json module in the Lua state as a global,
// consistent with obicontext, BioSequence, and http.
//
// Exposes:
//
//	json.encode(data)   → string         (on success)
//	json.encode(data)   → nil, err       (on error)
//	json.decode(string) → value          (on success)
//	json.decode(string) → nil, err       (on error)
func RegisterJSON(luaState *lua.LState) {
	table := luaState.NewTable()
	luaState.SetField(table, "encode", luaState.NewFunction(luaJSONEncode))
	luaState.SetField(table, "decode", luaState.NewFunction(luaJSONDecode))
	luaState.SetGlobal("json", table)
}

// luaJSONEncode implements json.encode(data) for Lua.
func luaJSONEncode(L *lua.LState) int {
	val := L.CheckAny(1)

	var goVal interface{}
	switch v := val.(type) {
	case *lua.LTable:
		goVal = Table2Interface(L, v)
	case lua.LString:
		goVal = string(v)
	case lua.LNumber:
		goVal = float64(v)
	case lua.LBool:
		goVal = bool(v)
	case *lua.LNilType:
		goVal = nil
	default:
		L.Push(lua.LNil)
		L.Push(lua.LString("json.encode: unsupported type"))
		return 2
	}

	b, err := json.Marshal(goVal)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LString(b))
	return 1
}

// luaJSONDecode implements json.decode(string) for Lua.
func luaJSONDecode(L *lua.LState) int {
	s := L.CheckString(1)

	var goVal interface{}
	if err := json.Unmarshal([]byte(s), &goVal); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	pushInterfaceToLua(L, goVal)
	return 1
}
