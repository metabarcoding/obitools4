# Taxonomic Classification via `TaxonomyClassifier`

The `obiseq` package provides a taxonomic classification mechanism through the `TaxonomyClassifier` function.

- **Purpose**: Constructs a reusable classifier for biological sequences based on taxonomic hierarchy.
- **Inputs**:
  - `taxonomicRank`: Target rank (e.g., `"species"`, `"genus"`).
  - `taxonomy`: Reference taxonomy (`*obitax.Taxonomy`), with fallback via `.OrDefault(true)`.
  - `abortOnMissing`: Boolean flag to enforce strict taxon resolution.

- **Core Logic**:
  - For each sequence, retrieves its `Taxon`, then drills down to the requested rank using `.TaxonAtRank()`.
  - If `abortOnMissing` is true, exits on failure to resolve the taxon or rank.
  - Internally maps `*TaxNode`s to integer codes for efficient storage/comparison.

- **Returned Object (`BioSequenceClassifier`)**:
  - `Code(sequence) int`: Assigns a unique integer code to the taxonomic assignment of a sequence.
  - `Value(code) string`: Returns the scientific name corresponding to a code.
  - `Reset()`: Reinitializes internal mappings (useful for batch processing).
  - `Clone() *BioSequenceClassifier`: Creates a fresh, identical classifier instance.

- **Design Rationale**:
  - Uses integer codes to avoid repeated string operations and enable fast indexing (e.g., for counting).
  - Supports both strict (`abortOnMissing=true`) and lenient classification modes.

This design enables scalable, efficient taxonomic profiling of sequencing datasets.
