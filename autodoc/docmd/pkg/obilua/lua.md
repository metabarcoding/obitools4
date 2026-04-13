# Obilua: Lua-Based Sequence Processing Framework

The `obilua` package provides a bridge between Go and the Lua scripting language for high-performance, parallelizable biological sequence processing. It enables users to write custom analysis logic in Lua while leveraging Go’s concurrency and I/O capabilities.

## Core Features

- **Lua Interpreter Initialization**: `NewInterpreter()` creates an isolated Lua state preloaded with Obi-specific types (`BioSequence`, etc.).
- **Compilation Support**: `Compile()` and `CompileScript()` parse and compile Lua code into efficient function prototypes.
- **Worker Conversion**: `LuaWorker(proto)` wraps a compiled Lua script as a Go-compatible `SeqWorker`, allowing seamless integration into sequence pipelines.
- **Pipeline Integration**: 
  - `LuaProcessor()` executes a Lua script over an iterator of sequences using configurable parallelism.
  - It supports optional `begin()` and `finish()` hook functions in Lua for initialization/cleanup.
  - Errors can be handled either by halting (`breakOnError=true`) or logging warnings.

- **Pipeable Interface**: 
  - `LuaPipe()` and `LuaScriptPipe()` expose Lua scripts as reusable, chainable pipeline stages (`obiiter.Pipeable`), supporting both inline programs and external `.lua` files.

## Lua API Contract

Scripts must define a global `worker(sequence)` function returning either:
- A single `BioSequence`
- A list (`BioSequenceSlice`)
Or return nothing (interpreted as filtered out).

Optionally, `begin()` and `finish()` functions may be defined for lifecycle management.

## Parallel Execution

Uses Go routines to run multiple workers concurrently, with batched input and output management. Default worker count falls back to system-wide parallelism settings if `nworkers ≤ 0`.

## Logging & Error Handling

Uses Logrus for structured logging; fatal errors are logged during setup, while runtime issues respect the `breakOnError` flag.
