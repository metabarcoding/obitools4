# Newick Format Export Functionality in `obiformats`

This Go package provides utilities to export taxonomic data into the **Newick format**, a standard for representing phylogenetic trees.

## Core Components

- `Tree`: A struct modeling a node in a Newick tree, containing:
  - `Children`: list of child nodes (nested trees),
  - `TaxNode`: reference to a taxonomic entry (`obitax.TaxNode`),
  - `Length`: optional branch length (evolutionary distance).

- **`Newick()` methods**:
  - `Tree.Newick(...)`: Recursively generates a Newick string for the subtree.
    Supports optional annotations: `scientific_name`, `taxid` (with `'@'` for rank), and branch lengths.
  - Package-level `Newick(...)`: Converts a full taxon set into a Newick tree string using the root node from `taxa.Sort().Get(0)`.

- **Writing Functions**:
  - `WriteNewick(...)`: Asynchronously writes the Newick representation to any `io.WriteCloser`.
    - Accepts an iterator over taxa (`*obitax.ITaxon`).
    - Validates single-taxonomy input.
    - Applies compression (via `obiutils.CompressStream`) if configured via options (`WithOption`).
  - `WriteNewickToFile(...)`: Convenience wrapper to write directly to a file.
  - `WriteNewickToStdout(...)`: Outputs Newick tree to standard output.

## Configuration Options

Options (e.g., `WithScientificName`, `WithTaxid`, `WithRank`) control annotation content and behavior (e.g., file closing, compression).

## Semantic Summary

The module enables **conversion of hierarchical taxonomic datasets into structured Newick trees**, supporting rich node labeling for downstream phylogenetic or bioinformatic tools.
