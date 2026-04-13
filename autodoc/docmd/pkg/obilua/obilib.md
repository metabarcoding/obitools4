# Obilib Module Overview

The `obilua` package provides Lua bindings for core OBIL (Ontology-Based Information Library) functionality, enabling scripting and extension of ontological data processing within a Lua environment.

## Core Components

- **`RegisterObilib(luaState *lua.LState)`**  
  Main registration function; initializes and exposes OBIL modules to a given Lua state.

- **`RegisterObiSeq(luaState *lua.LState)`**  
  Registers sequence-related operations (e.g., parsing, manipulation, and analysis of biological sequences like DNA/RNA/proteins).

- **`RegisterObiTaxonomy(luaState *lua.LState)`**  
  Registers taxonomy utilities (e.g., classification, lineage lookup, and hierarchical navigation of taxonomic trees).

## Semantic Capabilities

- Enables *semantic querying* over structured biological data via Lua scripts.
- Supports integration of ontological reasoning (e.g., using GO, NCBI Taxonomy) in dynamic workflows.
- Provides extensibility: new modules can be added by implementing `Register*` functions.

## Design Principles

- Minimal, non-intrusive API: only exposes essential high-level operations.
- Leverages `gopher-lua` for seamless interoperability between Go and Lua.

## Use Cases

- Custom annotation pipelines in bioinformatics.
- Interactive exploration of ontologies and sequences (e.g., via REPL or embedded Lua engines).
