# Taxonomic Annotation Features in `obiseq` Package

This package provides semantic taxonomic annotation capabilities for biological sequences (`BioSequence`). It integrates with a taxonomy database to assign, retrieve, and manage taxonomic identifiers (taxids) and related metadata.

## Core Functions

- **`Taxid()`**: Retrieves the taxonomic ID as a string (e.g., `"12345"` or `"NA"`), supporting multiple internal representations (`string`, `int`, `float64`). Returns `"NA"` if no taxid is set.

- **`Taxon(taxonomy)`**: Returns the corresponding `*obitax.Taxon` object, or `nil` if taxid is `"NA"`.

- **`SetTaxid(taxid, rank...)`**: Assigns a taxonomic ID to the sequence. Validates against default taxonomy; handles aliases and errors based on configuration flags (`FailOnTaxonomy`, `UpdateTaxid`). Optionally stores taxid under a custom rank (e.g., `"genus_taxid"`).

- **`SetTaxon(taxon, rank...)`**: Assigns a `*obitax.Taxon` object directly; stores its string representation as taxid.

## Rank-Specific Annotation

- **`SetTaxonAtRank(taxonomy, rank)`**: Annotates the sequence with taxid and scientific name at a specified Linnaean rank (e.g., `"species"`, `"genus"`). Sets two attributes: `rank_taxid` and `rank_name`. Returns the taxon at that rank (or `nil`).

- **Convenience wrappers**:
  - `SetSpecies(...)`
  - `SetGenus(...)`
  - `SetFamily(...)`  
    All delegate to `SetTaxonAtRank`.

## Taxonomic Path & Metadata

- **`SetPath(taxonomy)`**: Computes and stores the full taxonomic lineage (from root to species) as a string slice under attribute `"taxonomic_path"`.

- **`Path()`**: Retrieves the stored taxonomic path; recomputes it if missing and a default taxonomy exists.

- **`SetScientificName(taxonomy)`**: Stores the sequence’s species-level scientific name under `"scientific_name"`.

- **`SetTaxonomicRank(taxonomy)`**: Stores the taxon’s rank (e.g., `"species"`, `"genus"`) under `"taxonomic_rank"`.

## Error Handling & Configuration

- Uses `logrus` and custom logging (`obilog`) for warnings/errors.
- Behavior on taxonomy mismatches (e.g., unknown taxid, alias) is configurable via `obidefault` settings.
- Ensures type consistency: taxid must be string, int, or float; invalid types trigger fatal errors.

All methods are designed for seamless integration into bioinformatics pipelines, enabling robust taxonomic profiling of sequencing data.
