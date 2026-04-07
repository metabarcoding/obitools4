# `obilua` Module: Lua-Accessible Shared Context with Thread Safety

This Go package exposes a thread-safe, shared key-value context to Lua scripts via the Gopher-Lua interpreter.

## Core Features

- **Global `obicontext` Table**: Registered in Lua with the following methods:
  - `obicontext.item(key [, value])`:  
    Get or set a context variable. Supports types: `bool`, number, string, tables (converted via helper), and user data.
  - `obicontext.lock()`: Acquire exclusive lock on the context (blocking).
  - `obicontext.unlock()`: Release the global lock.
  - `obicontext.trylock()`: Attempt to acquire non-blocking lock; returns boolean success.
  - `obicontext.inc(key)` / `dec(key)`: Atomically increment/decrement numeric values (float64 only), with lock protection.

## Thread Safety

- Uses `sync.Mutex` for serializing write operations (e.g., inc/dec, lock/unlock).
- `sync.Map` for concurrent-safe read/write of key-value pairs.
- Critical sections (e.g., increment/decrement) are explicitly wrapped with locks to ensure atomicity.

## Lua Integration

- Values stored in the context persist across script calls.
- Type coercion is handled explicitly: Lua types map directly to Go equivalents, with fallback logging on unsupported types.
- Errors (e.g., incrementing non-number) trigger fatal logs—suitable for controlled environments.

## Use Case

Ideal for embedding Lua logic in Go applications requiring shared state (e.g., config, counters), with explicit locking for race-free updates.
