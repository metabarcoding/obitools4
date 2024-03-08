package obilua

import (
	log "github.com/sirupsen/logrus"

	lua "github.com/yuin/gopher-lua"
)

// pushInterfaceToLua converts a Go interface{} value to an equivalent Lua value and pushes it onto the stack.
//
// L *lua.LState: the Lua state onto which the value will be pushed.
// val interface{}: the Go interface value to be converted and pushed. This can be a basic type such as string, bool, int, float64,
// or slices and maps of these basic types. Custom complex types will be converted to userdata with a predefined metatable.
//
// No return values. This function operates directly on the Lua state stack.
func pushInterfaceToLua(L *lua.LState, val interface{}) {
	switch v := val.(type) {
	case string:
		L.Push(lua.LString(v))
	case bool:
		L.Push(lua.LBool(v))
	case int:
		L.Push(lua.LNumber(v))
	case float64:
		L.Push(lua.LNumber(v))
	// Add other cases as needed for different types
	case map[string]int:
		pushMapStringIntToLua(L, v)
	case map[string]string:
		pushMapStringStringToLua(L, v)
	case map[string]bool:
		pushMapStringBoolToLua(L, v)
	case map[string]float64:
		pushMapStringFloat64ToLua(L, v)
	case map[string]interface{}:
		pushMapStringInterfaceToLua(L, v)
	case []string:
		pushSliceStringToLua(L, v)
	case []int:
		pushSliceIntToLua(L, v)
	case []float64:
		pushSliceFloat64ToLua(L, v)
	case []bool:
		pushSliceBoolToLua(L, v)
	case nil:
		L.Push(lua.LNil)
	default:
		log.Fatalf("Cannot deal with value %v", val)
	}
}

func pushMapStringInterfaceToLua(L *lua.LState, m map[string]interface{}) {
	// Create a new Lua table
	luaTable := L.NewTable()
	// Iterate over the Go map and set the key-value pairs in the Lua table
	for key, value := range m {
		switch v := value.(type) {
		case int:
			luaTable.RawSetString(key, lua.LNumber(v))
		case float64:
			luaTable.RawSetString(key, lua.LNumber(v))
		case bool:
			luaTable.RawSetString(key, lua.LBool(v))
		case string:
			luaTable.RawSetString(key, lua.LString(v))
		default:
			log.Fatalf("Doesn't deal with map containing value %v", v)
		}
	}

	// Push the Lua table onto the stack
	L.Push(luaTable)
}

// pushMapStringIntToLua creates a new Lua table and iterates over the Go map to set key-value pairs in the Lua table. It then pushes the Lua table onto the stack.
//
// L *lua.LState - the Lua state
// m map[string]int - the Go map containing string to int key-value pairs
func pushMapStringIntToLua(L *lua.LState, m map[string]int) {
	// Create a new Lua table
	luaTable := L.NewTable()

	// Iterate over the Go map and set the key-value pairs in the Lua table
	for key, value := range m {
		L.SetTable(luaTable, lua.LString(key), lua.LNumber(value))
	}

	// Push the Lua table onto the stack
	L.Push(luaTable)
}

// pushMapStringStringToLua creates a new Lua table and sets key-value pairs from the Go map, then pushes the Lua table onto the stack.
//
// L *lua.LState, m map[string]string. No return value.
func pushMapStringStringToLua(L *lua.LState, m map[string]string) {
	// Create a new Lua table
	luaTable := L.NewTable()

	// Iterate over the Go map and set the key-value pairs in the Lua table
	for key, value := range m {
		L.SetTable(luaTable, lua.LString(key), lua.LString(value))
	}

	// Push the Lua table onto the stack
	L.Push(luaTable)
}

// pushMapStringBoolToLua creates a new Lua table, iterates over the Go map, sets the key-value pairs in the Lua table, and then pushes the Lua table onto the stack.
//
// Parameters:
//
//	L *lua.LState - the Lua state
//	m map[string]bool - the Go map
//
// Return type(s): None
func pushMapStringBoolToLua(L *lua.LState, m map[string]bool) {
	// Create a new Lua table
	luaTable := L.NewTable()

	// Iterate over the Go map and set the key-value pairs in the Lua table
	for key, value := range m {
		L.SetTable(luaTable, lua.LString(key), lua.LBool(value))
	}

	// Push the Lua table onto the stack
	L.Push(luaTable)
}

// pushMapStringFloat64ToLua pushes a map of string-float64 pairs to a Lua table on the stack.
//
// L *lua.LState - the Lua state
// m map[string]float64 - the map to be pushed to Lua
func pushMapStringFloat64ToLua(L *lua.LState, m map[string]float64) {
	// Create a new Lua table
	luaTable := L.NewTable()

	// Iterate over the Go map and set the key-value pairs in the Lua table
	for key, value := range m {
		// Use lua.LNumber since Lua does not differentiate between float and int
		L.SetTable(luaTable, lua.LString(key), lua.LNumber(value))
	}

	// Push the Lua table onto the stack
	L.Push(luaTable)
}

// pushSliceIntToLua creates a new Lua table and sets the elements of a Go slice in the Lua table. Then, it pushes the Lua table onto the stack.
//
// L *lua.LState, slice []int
// None
func pushSliceIntToLua(L *lua.LState, slice []int) {
	// Create a new Lua table
	luaTable := L.NewTable()

	// Iterate over the Go slice and set the elements in the Lua table
	for _, value := range slice {
		// Append the value to the Lua table
		// Lua is 1-indexed, so we use the length of the table + 1 as the next index
		luaTable.Append(lua.LNumber(value))
	}

	// Push the Lua table onto the stack
	L.Push(luaTable)
}

// pushSliceStringToLua creates a new Lua table and sets the elements in the table from the given Go slice. It then pushes the Lua table onto the stack.
//
// L *lua.LState - The Lua state
// slice []string - The Go slice of strings
func pushSliceStringToLua(L *lua.LState, slice []string) {
	// Create a new Lua table
	luaTable := L.NewTable()

	// Iterate over the Go slice and set the elements in the Lua table
	for _, value := range slice {
		// Append the value to the Lua table
		luaTable.Append(lua.LString(value))
	}

	// Push the Lua table onto the stack
	L.Push(luaTable)
}

// pushSliceBoolToLua creates a new Lua table and pushes the boolean values from the given slice onto the Lua stack.
//
// L *lua.LState - the Lua state
// slice []bool - the Go slice containing boolean values
func pushSliceBoolToLua(L *lua.LState, slice []bool) {
	// Create a new Lua table
	luaTable := L.NewTable()

	// Iterate over the Go slice and insert each boolean into the Lua table
	for _, value := range slice {
		// Lua is 1-indexed, so we use the length of the table + 1 as the next index
		luaTable.Append(lua.LBool(value))
	}

	// Push the Lua table onto the stack
	L.Push(luaTable)
}

// pushSliceFloat64ToLua creates a new Lua table and pushes it onto the stack.
//
// L *lua.LState - the Lua state
// slice []float64 - the Go slice to be inserted into the Lua table
func pushSliceFloat64ToLua(L *lua.LState, slice []float64) {
	// Create a new Lua table
	luaTable := L.NewTable()

	// Iterate over the Go slice and insert each float64 into the Lua table
	for _, value := range slice {
		// Lua is 1-indexed, so we append the value to the Lua table
		luaTable.Append(lua.LNumber(value))
	}

	// Push the Lua table onto the stack
	L.Push(luaTable)
}
