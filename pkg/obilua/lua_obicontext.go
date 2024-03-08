package obilua

import (
	"sync"

	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
)

var __lua_obicontext = &sync.Map{}
var __lua_obicontext_lock = &sync.Mutex{}

func RegisterObiContext(luaState *lua.LState) {

	table := luaState.NewTable()
	luaState.SetField(table, "item", luaState.NewFunction(obicontextGetSet))
	luaState.SetField(table, "lock", luaState.NewFunction(obicontextLock))
	luaState.SetField(table, "unlock", luaState.NewFunction(obicontextUnlock))
	luaState.SetField(table, "trylock", luaState.NewFunction(obicontextTrylock))
	luaState.SetField(table, "inc", luaState.NewFunction(obicontextInc))
	luaState.SetField(table, "dec", luaState.NewFunction(obicontextDec))

	luaState.SetGlobal("obicontext", table)
}

func obicontextGetSet(interpreter *lua.LState) int {
	key := interpreter.CheckString(1)

	if interpreter.GetTop() == 2 {
		value := interpreter.CheckAny(2)

		switch val := value.(type) {
		case lua.LBool:
			__lua_obicontext.Store(key, bool(val))
		case lua.LNumber:
			__lua_obicontext.Store(key, float64(val))
		case lua.LString:
			__lua_obicontext.Store(key, string(val))
		case *lua.LTable:
			__lua_obicontext.Store(key, Table2Interface(interpreter, val))
		default:
			log.Fatalf("Cannot store values of type %s in the obicontext", value.Type().String())
		}

		return 0

	}

	if value, ok := __lua_obicontext.Load(key); ok {
		pushInterfaceToLua(interpreter, value)
	} else {
		interpreter.Push(lua.LNil)
	}

	return 1
}

func obicontextInc(interpreter *lua.LState) int {
	key := interpreter.CheckString(1)
	__lua_obicontext_lock.Lock()

	if value, ok := __lua_obicontext.Load(key); ok {
		if v, ok := value.(float64); ok {
			v++
			__lua_obicontext.Store(key, v)
			__lua_obicontext_lock.Unlock()
			interpreter.Push(lua.LNumber(v))
			return 1
		}
	}

	__lua_obicontext_lock.Unlock()
	log.Fatalf("Cannot increment item %s", key)

	return 0
}

func obicontextDec(interpreter *lua.LState) int {
	key := interpreter.CheckString(1)
	__lua_obicontext_lock.Lock()
	defer __lua_obicontext_lock.Unlock()

	if value, ok := __lua_obicontext.Load(key); ok {
		if v, ok := value.(float64); ok {
			v--
			__lua_obicontext.Store(key, v)
			interpreter.Push(lua.LNumber(v))
			return 1
		}
	}

	log.Fatalf("Cannot decrement item %s", key)

	return 0
}

func obicontextLock(interpreter *lua.LState) int {

	__lua_obicontext_lock.Lock()

	return 0
}

func obicontextUnlock(interpreter *lua.LState) int {

	__lua_obicontext_lock.Unlock()

	return 0
}

func obicontextTrylock(interpreter *lua.LState) int {

	result := __lua_obicontext_lock.TryLock()

	interpreter.Push(lua.LBool(result))
	return 1
}
