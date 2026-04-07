# `obidefault` Package: Configuration State Management

This Go package provides a centralized, thread-safe(ish) configuration layer for taxonomy-related settings in the OBIDMS (Open Biological and Biomedical Data Management System) framework. It exposes simple getters, setters, and pointer accessors for four core boolean/string flags that control how taxonomic identifiers (taxids) are handled during data processing.

## Core Configuration Flags

- `__taxonomy__`: Stores the currently selected taxonomy (e.g., `"NCBI"`, `"UNIPROT"`).  
- `__alternative_name__`: Enables/disables use of alternative taxonomic names (e.g., synonyms).  
- `__fail_on_taxonomy__`: If true, processing halts on taxonomy mismatches/errors.  
- `__update_taxid__`: If true, taxids are auto-updated to current NCBI/DB versions.  
- `__raw_taxid__`: If true, raw (unprocessed) taxids are preserved instead of normalized.

## Public API

- **Getters**: `UseRawTaxids()`, `SelectedTaxonomy()`, `HasSelectedTaxonomy()`, etc., return current values.  
- **Pointer Accessors**: e.g., `SelectedTaxonomyPtr()` returns a pointer for direct mutation (advanced use).  
- **Setters**: `SetSelectedTaxonomy()`, `SetAlternativeNamesSelected()`, etc., update state.

## Use Case

Typically used at application startup to configure global behavior (e.g., `SetSelectedTaxonomy("NCBI")`, `SetUpdateTaxid(true)`), then referenced by downstream modules during data import, validation, or mapping. Minimalist and explicit—no external dependencies.
