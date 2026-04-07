# `ExtractTaxonomy` Function — Semantic Description

The `ExtractTaxonomy` method is a core utility in the `obiiter` package, designed to aggregate taxonomic information across biological sequences processed by an iterator.

- **Input**:  
  - A pointer to `IBioSequence`, representing a sequence iterator over biological data.  
  - A boolean flag `seqAsTaxa`: if true, each full sequence is treated as a single taxonomic unit; otherwise, individual elements within slices are processed separately.

- **Process**:  
  - Iterates through all sequences via `iterator.Next()` and retrieves each current slice using `Get().Slice()`.  
  - For every slice, it calls the underlying `.ExtractTaxonomy()` method (from `obitax`), progressively building or updating a shared `*obitax.Taxonomy` object.  
  - Stops and returns immediately upon encountering the first error during taxonomy extraction.

- **Output**:  
  - Returns a fully populated `*obitax.Taxonomy` object (or partial result if early failure occurs).  
  - Returns `nil` error on success; otherwise, returns the first encountered error.

- **Semantic Role**:  
  Enables scalable taxonomic profiling of high-throughput sequencing data by delegating per-slice extraction logic to the `obitax` module, while ensuring robust iteration and error handling.

- **Dependencies**:  
  Relies on `obitax.Taxonomy` for structured taxonomic representation and assumes slices implement the `.ExtractTaxonomy()` interface.

This function exemplifies a *map-reduce*-style pattern: mapping taxonomy extraction over slices, and reducing results into a unified taxonomic summary.
