# `obilua.Mutex`: Thread-Safe Synchronization in Lua via Go's sync.Mutex

This package exposes **Go’s `sync.Mutex`** to the Lua environment using [gopher-lua](https://github.com/yuin/gopher-lua), enabling safe concurrent access from Lua scripts.

## Key Features

- **Custom userdata type**: Registers a new metatable `"Mutex"` in the Lua state.
- **Constructor function**:
  - ` Mutex.new() → mutex userdata`  
    Creates and returns a new Go-backed mutex instance.
- **Instance methods**:
  - `mutex:lock()` — Acquires the lock (blocks until available).
  - `mutex:unlock()` — Releases the lock.
- **Type safety**: Validates that only valid mutex userdatas are passed to `lock`/`unlock`.
- **Integration**: Designed for embedding Lua in Go applications requiring synchronization (e.g., multi-threaded scripting).

## Usage Example

```lua
local m = Mutex.new()
m:lock()   -- Acquire lock (safe across goroutines)
-- critical section
m:unlock()
```

## Implementation Notes

- Mutex state is stored in a Go `*sync.Mutex` inside Lua userdata.
- No reference counting or finalizers — user must manually manage lock/unlock lifecycle to avoid deadlocks.
- Thread-safe *from Go side only*; Lua calls must respect goroutine safety (e.g., avoid calling from multiple VMs concurrently).
