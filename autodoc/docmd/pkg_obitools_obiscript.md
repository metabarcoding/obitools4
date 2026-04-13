# `obiscript` Package: CLI Scriptable Processing Pipeline

The `obiscript` package provides a high-level, modular interface for embedding custom Lua scripts into OBITools4’s sequence processing pipelines. It enables users to define bioinformatics workflows using a lightweight, embeddable scripting language—without sacrificing performance or composability.

## Public API Overview

### `CLIScriptPipeline() Pipeable`
Returns a reusable, parallelized pipeline stage that executes user-provided Lua scripts on sequence data. Internally uses `obilua.LuaScriptPipe()` with parallelism enabled (default worker count from `obidefault.ParallelWorkers`). Accepts a script path via the pipeline configuration (typically set through `CLIScriptFilename()` in CLI usage). Designed to integrate seamlessly with other pipeable stages from the `obiiter` framework.

### Script Lifecycle Hooks (Exposed via Lua API)
The embedded ObiLua runtime expects the user script to define three optional functions:

- **`begin()`**  
  Called once before processing any sequences. Used for initialization (e.g., counters, file handles). Optional.

- **`worker(sequence)`**  
  Invoked for each input sequence. Provides full access to metadata via `sequence:attribute(name, value?)`, supports in-place modification of tags/IDs, and allows interaction with global context (`obicontext`) for cross-sequence state management.

- **`finish()`**  
  Called after all sequences have been processed. Typically used to output summary statistics or cleanup resources.

### CLI Integration

- **`--script FILE`, `-S FILE`**  
  Specifies the Lua script to execute. The file must exist and be syntactically valid.

- **`--template`**  
  Outputs a minimal, self-contained Lua script template to stdout. Includes stubs for `begin()`, `worker(...)`, and `finish()` with inline documentation.

- **Shared Options**  
  Reuses configuration sets from core OBITools4 modules:
    - `obiconvert.DataIOOptionSet`: input/output format, file paths.
    - `obigrep.SequenceSelectionOptionSet`: filtering/sorting logic.

## Semantic Role

`obiscript` abstracts the complexity of embedding and orchestrating Lua scripts in a streaming, parallelizable context. It decouples *workflow logic* (Lua) from *pipeline orchestration* (Go), enabling:
- Rapid prototyping of NGS processing steps.
- Custom annotation, filtering, assembly, or reporting without recompilation.
- Consistent CLI behavior across OBITools4 tools.

## Use Cases

| Scenario | Example |
|---------|---------|
| Read filtering + renaming | Filter low-quality reads and prepend sample ID to sequence names |
| Annotation injection | Add UMI or barcode info from external metadata file per read |
| Summary reporting | Count reads passing filters, write stats to log at end |

> *Designed for extensibility: users extend functionality by writing Lua, not Go.*
