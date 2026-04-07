# NGSFilter Configuration Parser — Semantic Overview

This Go package (`obiformats`) provides robust parsing and validation of NGS (Next-Generation Sequencing) filter configurations used in the OBITools4 ecosystem. It supports two legacy and modern formats: a line-based text format (`ReadOldNGSFilter`) and CSV-based configuration files with parameter headers.

## Core Functionality

- **Format Detection**:  
  `OBIMimeNGSFilterTypeGuesser` detects MIME type using content sniffing (via [`mimetype`](https://github.com/gabriel-vasile/mimetype)), distinguishing between `text/csv`, custom `text/ngsfilter-csv`, and plain text.  
  A heuristic CSV detector (`NGSFilterCsvDetector`) validates structure (consistent column count, non-empty rows).

- **Dual Input Parsing**:
  - `ReadOldNGSFilter`: Parses line-based config files (e.g., lines like `"EXP1@SAMPLE1:TAGFWD-TAGREV primer_f primer_r"`), supporting:
    - Primer pairs (`forward`, `reverse`)
    - Tag pairs (with optional `-` for untagged direction)
    - Experiment/sample metadata
    - OBIFeatures annotations (via `ParseOBIFeatures`)
  - `ReadCSVNGSFilter`: Parses structured CSV files with mandatory columns:  
    `"experiment"`, `"sample"`, `"sample_tag"`, `"forward_primer"`, `"reverse_primer"`  
    Additional columns are stored as annotations.

- **Parameter Configuration**:
  A rich set of `@param` lines (in CSV or legacy format) configures global/primers-specific settings:
  - `spacer`, `forward_spacer`, `reverse_spacer`: Tag-primer spacing (bp)
  - `tag_delimiter` / directional variants: Symbol separating tags in sequences
  - `matching`: Tag matching algorithm (e.g., exact, fuzzy)
  - Error tolerance:  
    `primer_mismatches`, `forward_mismatches`, `reverse_mismatches` (max mismatches)  
    `tag_indels`, `forward_tag_indels`, etc. (allow indel errors)
  - Indel handling:  
    `indels` / directional variants (`true/false`) to enable/disable indels in primer matching

- **Validation & Integrity Checks**:
  - `CheckPrimerUnicity`: Ensures each primer pair is defined only once.
  - Duplicate tag-pair detection per marker (error on reuse).
  - Strict column/field validation with informative error messages.

- **Logging & Observability**:
  Uses `logrus` for detailed info/warnings (e.g., parameter application, skipped unknown params).

## Design Highlights

- **Extensibility**: New parameters can be added via `library_parameter` map.
- **Robustness**: Handles BOM, line continuation (`ReadLines`), CSV quirks (lazy quotes, comments).
- **Semantic Clarity**: Separates *data* (samples/markers/tags) from *configuration* (parameters).
- **Integration Ready**: Returns a validated `obingslibrary.NGSLibrary` ready for downstream processing.

> **Use Case**: Enables reproducible, metadata-rich NGS filtering setups in metabarcoding workflows.
