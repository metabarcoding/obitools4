# Semantic Description of `obilua` Package

This Go package provides utilities for converting Lua tables—used in a Gopher-Lua environment—to native Go data structures.

- **`Table2Interface`**:  
  Converts a Lua `*lua.LTable` into either:
  - A Go slice (`[]interface{}`) if the table is array-like (keys are numeric, starting at 1), preserving order and type coercion (`nil`, `bool`, `float64`, `string`).
  - A Go map (`map[string]interface{}`) if the table contains string keys (i.e., a hash/dictionary).

- **`Table2ByteSlice`**:  
  Specifically converts an array-like Lua table into a `[]byte`, assuming all values are numeric and ≤ 255.  
  - Fails with a fatal log if non-numeric or out-of-range values are encountered.
  - Also fails fatally for hash-like (non-array) tables.

- **Key Design Notes**:
  - Type coercion is explicit and safe: only `LTNil`, `LTBool`, `LTNumber`, `LTString` are supported.
  - Array detection relies on key type: if *all* keys are `LNumber`, the table is treated as an array.
  - Uses [`logrus`](https://github.com/sirupsen/logrus) for fatal error reporting.
  - No dependency on external serialization (e.g., JSON); conversions are direct and lightweight.

- **Use Cases**:
  - Bridging Lua scripting layers with Go backends (e.g., embedded config parsing, plugin systems).
  - Efficiently extracting structured data from Lua state into idiomatic Go types.

> ⚠️ **Limitations**:  
> - No support for nested tables or custom types.  
> - Array indexing assumes 1-based Lua semantics (converted to 0-indexed Go slices).  
> - No error handling: misuse triggers `log.Fatalf`.
