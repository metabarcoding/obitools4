# Semantic Description of `obingslibrary` Marker Module

The `Marker` struct defines a molecular biology primer pair (forward/reverse) for PCR-based sample demultiplexing in high-throughput sequencing workflows. It supports flexible configuration of primer binding, tag (barcode) extraction, mismatch tolerance, and indel handling.

## Core Functionalities

- **Primer Pattern Compilation**:  
  `Compile()` and `Compile2()` initialize forward/reverse primer patterns using the underlying `obiapat.ApatPattern`, including reverse-complement variants (`cforward`, `creverse`). They accept parameters for maximum error tolerance and indel allowance.

- **Sequence Matching & Demultiplexing**:  
  `Match()` scans a given sequence (`BioSequence`) for primer binding sites. It prioritizes forward-primer detection, then falls back to reverse if needed. For each match:
  - Extracts primer region and adjacent tag (barcode).
  - Computes mismatches.
  - Links to a pre-registered `PCR` object via the tag pair (`TagPair`) key in internal map.

- **Sample Registration & Lookup**:  
  `GetPCR()` retrieves or registers a new PCR reaction entry indexed by forward/reverse tag pair (case-insensitive). Enables tracking of sample-specific amplification data.

- **Tag Length Validation**:  
  `CheckTagLength()` ensures all registered tags have uniform length for both directions; otherwise, returns an error.

- **Configurable Parameters**:  
  Supports tuning of:
  - Tag lengths (`Forward_tag_length`, `Reverse_tag_length`)
  - Spacer between tag and primer (`SetTagSpacer()`)
  - Delimiter for tag-primer boundary (e.g., `a`, `c`, `g`, `t` or none via `'0'`)
  - Allowed mismatches and indels per primer (`SetAllowedMismatch()`, `SetTagIndels()`)
  - Matching strategy: `"strict"` (exact), `"hamming"`, or `"indel"`

- **Matching Strategy Enforcement**:  
  `SetForward/ReverseMatching()` validates and sets matching modes; invalid values raise errors.

## Design Highlights

- Uses `log.Fatalf` for critical configuration failures (e.g., invalid delimiter).
- Leverages reference-counted sequences (`Recycle()`) for memory efficiency.
- Prioritizes forward primer match but gracefully handles reverse orientation.
- Fully supports case-insensitive tag comparison and normalization.

This module serves as the core engine for sample assignment in amplicon-based NGS pipelines, balancing sensitivity (via error/indel tolerance) and specificity (through tag uniqueness).
