# `obiphylo` Package: Semantic Description

The `obiphylo` package provides a minimal yet expressive data structure and utilities for representing **phylogenetic trees** in Go, prioritizing simplicity, extensibility, and interoperability with standard phylogenetic formats.

## Core Type: `PhyloNode`

Represents a node in a phylogenetic tree—either an operational taxonomic unit (leaf) or an internal branching point.

### Public Fields
- `Name string`: Optional identifier for the node (e.g., species name, OTU label). May be empty.
- `Children map[*PhyloNode]float64`: Maps child nodes to their associated **branch lengths** (evolutionary distances). Supports `NaN` for unspecified or unmeasured branches.
- `Attributes map[string]any`: A flexible key-value store for arbitrary metadata (e.g., bootstrap values, posterior probabilities, geographic origin). Values may be of any type.

> ⚠️ *All fields are exported for direct read/write access, but users should prefer the provided methods to ensure consistency (e.g., `AddChild`, `SetAttribute`).*

## Public Methods

### Construction & Mutation
- **`NewPhyloNode(name string) *PhyloNode`**  
  Instantiates a new node with optional name. Initializes `Children` and `Attributes` as empty maps.

- **`AddChild(child *PhyloNode, distance float64)`**  
  Appends a child node to the current one with specified branch length. If `distance` is `NaN`, it is stored as-is (and omitted in Newick output).  
  → *Enables incremental tree building from leaves to root.*

- **`SetAttribute(key string, value any)`**  
  Stores or updates a metadata entry on the node. Overwrites existing keys.

- **`GetAttribute(key string) (any, bool)`**  
  Retrieves a metadata value and reports presence via boolean. Returns zero `value` if key absent.

### Tree Serialization
- **`Newick(level int) string`**  
  Recursively generates a Newick-formatted subtree rooted at the current node.  
  - Nodes without children appear as `Name` (or empty string if unnamed).  
  - Internal nodes are rendered with comma-separated children in parentheses.  
  - Branch lengths (`:distance`) appear *only if finite* (i.e., `!math.IsNaN(distance)`).  
  - Indentation (`level * "\t"`) improves human readability.  
  - Root-level calls (e.g., `root.Newick(0)`) append a final semicolon (`;`).  
  → *Designed for export to tools like RAxML, FigTree, or Iq-TREE.*

## Design Principles

- **Zero external dependencies**: Pure Go implementation.
- **Idiomatic efficiency**: Child lookup via `map` ensures O(1) average access.
- **Extensibility over rigidity**: Arbitrary metadata via `any` supports evolving annotation needs without API changes.
- **Format compliance**: Newick output adheres to widely accepted syntax (with optional branch lengths), enabling seamless integration with phylogenetic software ecosystems.

## Usage Example

```go
root := obiphylo.NewPhyloNode("Root")
leafA := obiphylo.NewPhyloNode("Species_A")
leafB := obiphylo.NewPhyloNode("Species_B")

root.AddChild(leafA, 1.2)
root.AddChild(leafB, math.NaN()) // distance omitted in output

leafA.SetAttribute("bootstrap", 95)
root.Newick(0) // → "\t(Species_A:1.2,Species_B);"
```
