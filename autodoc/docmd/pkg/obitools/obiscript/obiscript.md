# `obiscript` Package: CLI Script Pipeline

This Go module defines a high-level pipeline interface for executing Lua-based processing scripts within the OBITools4 ecosystem.

## Core Functionality

- **`CLIScriptPipeline()`**  
  Returns a `Pipeable` iterator pipeline configured to run user-provided Lua scripts via the command-line interface.

- **Implementation Details**  
  - Uses `obilua.LuaScriptPipe()` to instantiate a Lua-based processing stage.
    - Accepts the script filename (via `CLIScriptFilename()`).
    - Enables parallel execution (`true` flag) using default worker count from `obidefault.ParallelWorkers()`.

- **Integration**  
  - Built on top of the `obiiter` iterator framework, allowing composition with other pipeable operations.
  - Designed for CLI usage: expects a Lua script path (likely passed via `--script` or similar flag).

## Semantic Role

This function abstracts the setup of a *Lua-scriptable processing stage*—enabling users to inject custom filtering, annotation, transformation, or assembly logic in Lua while preserving parallelism and pipeline modularity.

## Use Case

Ideal for building modular, scriptable NGS data processing workflows (e.g., read filtering → annotation → consensus generation), where flexibility and performance are both required.
