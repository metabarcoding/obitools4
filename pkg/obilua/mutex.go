package obilua

import (
	lua "github.com/yuin/gopher-lua"

	"sync"
)

const luaMutexTypeName = "Mutex"

func registerMutexType(luaState *lua.LState) {
	mutexType := luaState.NewTypeMetatable(luaMutexTypeName)
	luaState.SetGlobal(luaMutexTypeName, mutexType)

	luaState.SetField(mutexType, "new", luaState.NewFunction(newMutex))

	luaState.SetField(mutexType, "__index",
		luaState.SetFuncs(luaState.NewTable(),
			mutexMethods))
}

func mutex2Lua(interpreter *lua.LState, mutex *sync.Mutex) lua.LValue {
	ud := interpreter.NewUserData()
	ud.Value = mutex
	interpreter.SetMetatable(ud, interpreter.GetTypeMetatable(luaMutexTypeName))

	return ud
}

func pushMutexToLua(luaState *lua.LState, mutex *sync.Mutex) {
	luaState.Push(mutex2Lua(luaState, mutex))
}
func newMutex(luaState *lua.LState) int {
	m := &sync.Mutex{}
	pushMutexToLua(luaState, m)
	return 1
}

var mutexMethods = map[string]lua.LGFunction{
	"lock":   mutexLock,
	"unlock": mutexUnlock,
}

func checkMutex(L *lua.LState) *sync.Mutex {
	ud := L.CheckUserData(1)
	if mutex, ok := ud.Value.(*sync.Mutex); ok {
		return mutex
	}
	L.ArgError(1, "Mutex expected")
	return nil
}

func mutexLock(L *lua.LState) int {
	mutex := checkMutex(L)
	mutex.Lock()
	return 0
}

func mutexUnlock(L *lua.LState) int {
	mutex := checkMutex(L)
	mutex.Unlock()
	return 0
}
