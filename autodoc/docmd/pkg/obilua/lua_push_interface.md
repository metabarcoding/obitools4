# Semantic Description of `obilua` Package

The `obilua` package provides utilities for **bi-directional data marshaling between Go and Lua**, specifically focusing on converting native Go values into equivalent `lua.LValue` types for use in a Lua state (`*lua.LState`). This enables Go applications to expose structured data (e.g., maps, slices) or synchronization primitives (`*sync.Mutex`) directly to Lua scripts.

## Core Functionality

- **`pushInterfaceToLua(L, val)`**:  
  Main dispatcher that inspects the type of a Go `interface{}` value and routes it to specialized conversion functions. Supported types include:
  - Basic scalar types: `string`, `bool`, `int`, `float64`
  - Collections:
    - Maps: `map[string]{string,int,bool,float64,interface{}}`
    - Slices/arrays: `[]{string,int,byte,float64,bool]interface{}}`
  - Special cases:
    - `nil` → Lua’s `LNil`
    - `*sync.Mutex` (via dedicated handler)

- **Type-Specific Pushers**:  
  Each helper function (`pushMapStringIntToLua`, `pushSliceBoolToLua`, etc.) constructs a new Lua table and populates it with converted elements using appropriate `lua.LValue` constructors (`LString`, `LNumber`, `LBool`).  
  - Maps are converted as associative tables (keyed by string).
  - Slices become indexed Lua arrays (`1..n`).

- **Generic Slice Support**:  
  `pushSliceNumericToLua[T]()` uses Go generics to handle numeric slices (`int`, `float64`, `byte`) uniformly.

## Design Notes

- **No reverse conversion** (Lua → Go) is included — only *pushing* to Lua.
- **Strict typing**: Unsupported types trigger a fatal log (`log.Fatalf`), enforcing explicit type handling.
- **Lua semantics respected**: Tables are 1-indexed, and numeric types map to `lua.LNumber`.

This package is ideal for embedding Lua in Go services where dynamic configuration, rule evaluation, or scripting requires safe and predictable data injection.
