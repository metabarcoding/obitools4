# ObiScript CLI: Scriptable Sequence Processing Framework

ObiScript provides a command-line interface for executing custom Lua scripts against biological sequence data within the OBITools4 ecosystem.

## Core Functionality

- **Script Execution (`--script` / `-S`)**
  - Accepts a path to a Lua script file.
  - The script is read and executed using the embedded ObiLua runtime.

- **Script Template Generation (`--template`)**
  - Outputs a minimal, executable script template to stdout.
  - Template defines `begin()`, `worker(sequence)`, and `finish()` lifecycle hooks.

- **Integration with OBITools4 Modules**
  - Reuses configuration options from `obiconvert` (data I/O, format handling).
  - Integrates sequence filtering/sorting via `obigrep.SequenceSelectionOptionSet`.

## Script Lifecycle

1. **`begin()`**  
   Initialize global state (e.g., counters, resources).

2. **`worker(sequence)`**  
   Process each sequence individually:
   - Access/modify metadata via `sequence:attribute(...)`.
   - Assign new IDs or enrich annotations.
   - Use global context (`obicontext`) for cross-sequence state.

3. **`finish()`**  
   Finalize and output summary (e.g., print counters).

## Example Workflow

A typical script increments a counter, updates sample metadata, and renames sequences — demonstrating extensible transformation logic without recompilation.

## Design Principles

- **Modularity**: Script behavior is decoupled from CLI logic.
- **Extensibility**: Lua scripting enables complex, user-defined pipelines.
- **Consistency**: Aligns with existing OBITools4 CLI conventions via shared option sets.

> *ObiScript bridges high-level bioinformatics workflows with low-level sequence manipulation via embedded Lua.*
