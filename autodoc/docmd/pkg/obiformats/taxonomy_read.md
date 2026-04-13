# Taxonomy Loading Module (`obiformats`)

This Go package provides semantic functionality to automatically detect and load taxonomic data from various file formats. It supports flexible, format-agnostic taxonomy ingestion via a unified interface.

## Core Features

1. **Format Detection**  
   - `DetectTaxonomyFormat(path)` identifies the taxonomy source format by inspecting file type (directory, MIME-type), filename patterns, or structure.
   - Supports:  
     • NCBI Taxdump (both directory and `.tar` archive)  
     • CSV files (`text/csv`)  
     • FASTA/FASTQ sequences (via `mimetype` detection)  

2. **Modular Loaders**  
   - Returns a typed `TaxonomyLoader` function, enabling deferred loading with configurable options (`onlysn`, `seqAsTaxa`).  
   - Each loader abstracts format-specific parsing (e.g., NCBI `nodes.dmp`, FASTA header taxonomy extraction).

3. **Sequence-Based Taxonomy Extraction**  
   - For sequence files (FASTA/FASTQ), taxonomy is inferred from headers or associated metadata, using `ExtractTaxonomy()`.

4. **Integration with OBITools Ecosystem**  
   - Leverages `obitax.Taxonomy` as the canonical output structure.  
   - Uses custom MIME-type registration (`obiutils.RegisterOBIMimeType()`) for robust detection of bioinformatics formats.

5. **Error Handling & Logging**  
   - Graceful failure with descriptive errors; informative logging via `logrus`.

## Usage Flow

```go
tax, err := LoadTaxonomy("path/to/data", onlysn=true, seqAsTaxa=false)
```

The module enables interoperability across taxonomic data sources in metabarcoding workflows.
