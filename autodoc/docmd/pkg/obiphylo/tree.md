# `obiphylo` Package: Semantic Description

The `obiphylo` package provides a minimal yet expressive data structure and utilities for representing **phylogenetic trees** in Go.

## Core Type: `PhyloNode`
- Represents a node (taxon or internal branch point) in a phylogeny.
- Fields:
  - `Name`: Optional label (e.g., species name, OTU ID).
  - `Children`: A map of child nodes to **branch lengths** (evolutionary distances).
  - `Attributes`: A flexible key-value store for metadata (e.g., bootstrap support, posterior probability).

## Key Functionalities
- **Tree Construction**:
  - `NewPhyloNode()`: Instantiates an empty node.
  - `AddChild(child, distance)`: Appends a child with associated branch length (supports NaN for unlabeled branches).
- **Metadata Access**:
  - `SetAttribute(key, value)` / `GetAttribute(key)`: Enables extensible node annotation.
  - Supports arbitrary types (via `any`), ideal for dynamic metadata.

## Output: Newick Format
- Recursive method `Newick(level int)` generates a **human-readable, standard phylogenetic tree string**:
  - Properly indented for readability.
  - Supports branch lengths (`:distance`) on edges (skips if `NaN`).
  - Terminates with semicolon (`;`) at root level.
- Designed for interoperability (e.g., export to tools like RAxML, FigTree).

## Design Notes
- Lightweight and dependency-free.
- Uses Go’s idiomatic maps for efficient child lookup (O(1) average).
- Recursive Newick generation ensures correct nesting and formatting.
