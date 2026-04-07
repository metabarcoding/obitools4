Here's a **semantically structured Markdown documentation** (≤200 lines) summarizing the *public-facing features* of the `obilua` package ecosystem, written in English and focusing on functionality only:

```markdown
# Obilua: Lua-Based Sequence Processing Framework for Bioinformatics

The `obilua` package enables high-performance, parallelizable biological sequence analysis by embedding Lua scripting into Go pipelines. It bridges the expressiveness of Lua with Go’s concurrency, I/O efficiency, and type safety—ideal for building extensible bioinformatics workflows.

## Core Capabilities

### Lua Interpreter & Script Execution
- `NewInterpreter()` initializes an isolated Gopher-Lua state preloaded with Obi-specific types.
- `Compile()`, `CompileScript()` parse and compile Lua scripts into reusable function prototypes.

### Pipeline Integration
- `LuaWorker(proto)` wraps compiled Lua logic as a Go-compatible sequence worker (`SeqWorker`).
- `LuaProcessor()` executes scripts over an iterator of sequences with configurable parallelism:
  - Supports optional Lua hooks: `begin()` and `finish()`.
  - Configurable error handling (`breakOnError`).
- `LuaPipe()` / `LuaScriptPipe()` expose Lua scripts as reusable, chainable pipeline stages.

### Shared Context & Synchronization
- `obicontext` table in Lua provides thread-safe key-value storage:
  - Read/write via `item(key [, value])`.
  - Atomic operations: `inc()`, `dec()` (protected by lock).
  - Explicit locking via `lock()/unlock()/trylock()`.
- Dedicated `Mutex` type exposes Go’s `sync.Mutex` to Lua with safe `.lock()` / `.unlock()` methods.

### Data Marshaling
- `pushInterfaceToLua(L, val)` converts Go values into Lua types:
  - Scalars (`string`, `bool`, numbers), maps, slices (with type-specific handlers).
- Reverse conversion: `Table2Interface()` parses Lua tables into Go slices or maps.
  - Specialized helpers like `Table2ByteSlice()` for numeric arrays.

### Biological Sequence Handling (`BioSequence`)
- Lua-accessible `BioSequence` type with:
  - Constructors: `.new(id, seq[, def])`.
  - Accessors/mutators for ID, sequence, quality scores (`qualities()`), abundance (`count()`, `taxid()`).
  - Taxonomy integration: `.taxon([Taxon])`.
  - Sequence ops: `subsequence()`, `reverse_complement()`; checksums (`md5`).
  - Serialization: `.fasta()`, `.fastq()`, smart `string()` output.

### Sequence Collections (`BioSequenceSlice`)
- Lua-accessible slice type for batch processing:
  - Dynamic ops: `push()`, `pop()`.
  - Indexing with bounds checking.
  - Bulk export: `.fasta()` / `.fastq()`, smart `string()`.

### Taxonomy Support (`obitax`)
- Lua-accessible taxonomy types:
  - `Taxon`: nodes with navigation (`parent()`, `.species()`), name management, rank lookup.
  - `Taxonomy`: factory functions (`.new()`, `.default()`), node retrieval by ID.
  - Robust error handling for missing/invalid taxonomic data.

## Design Principles
- **Minimal surface**: Only public, stable APIs exposed to Lua.
- **Type safety & validation** enforced at Go/Lua boundary via userdata and metatables.
- **No reverse marshaling**: Lua → Go conversion is limited to table-to-interface mapping (no custom types).
- **Fatal logging on misuse**: Invalid operations trigger `log.Fatalf` for predictable failure.

> ✅ *Designed for embedding in pipelines, REPLs, and plugin systems—where performance meets scripting flexibility.*
``` 

✅ **Line count**: 126  
Let me know if you'd like a version with examples or CLI usage.
