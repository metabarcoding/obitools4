# `obingslibrary`: High-Throughput Sequencing Demultiplexing Library

`obingslibrary` is a Go package for **sample assignment in amplicon-based NGS workflows**, using dual-indexed barcodes flanked by PCR primers. It enables robust, configurable demultiplexing of sequencing reads—even in the presence of errors or indels—by matching primer–tag patterns and assigning samples via tag lookup.

---

## Core Functionalities

### 1. **Primer & Tag Configuration**
- `Marker`: Defines a primer pair (forward/reverse), including:
  - Primer sequences (`Forward`, `Reverse`) and reverse-complement variants.
  - Tag specifications: lengths, spacers (e.g., `N` or fixed nucleotides), delimiters.
  - Mismatch/indel tolerance per direction (`SetAllowedMismatch`, `SetTagIndels`).
- **Compilation**:  
  - `Compile()` / `Compile2()`: Builds internal pattern indexes (via `obiapat.ApatPattern`) for fast, error-tolerant matching.
  - Supports `"strict"`, `"hamming"` (substitutions only), or `"indel"` (Levenshtein) matching modes.

### 2. **Sequence Matching & Demultiplexing**
- `Match(sequence)`: Scans a `BioSequence` for valid primer bindings:
  - Prioritizes forward-primer detection; falls back to reverse orientation.
  - Returns `DemultiplexMatch` with:
    - Primer positions, mismatches, orientation (`IsDirect`).
    - Barcode coordinates (`BarcodeStart`, `BarcodeEnd`) and validity flag.
- **Primer dimer detection**: If `BarcodeStart > BarcodeEnd`, the read is flagged as invalid.

### 3. **Tag Extraction & Annotation**
- `ExtractBarcode(sequence, inplace)`:
  - Extracts the barcode region between forward/reverse primers.
  - Reverse-complements if read is in reverse orientation (`IsDirect == false`).
  - Annotates the sequence with:
    - Primer names, positions, mismatches.
    - Sample/experiment info (if tag assignment succeeds).
    - Error messages (`Unassigned`, `NoMatch`, etc.).
- **Tag extraction strategies**:
  - `Fixed`: Fixed-length tags.
  - `Delimited`: Tags flanked by exact delimiters (e.g., `"NN"`).
  - `Rescue`: Tolerates indels in delimiter or tag boundaries.

### 4. **Sample Registration & Lookup**
- `GetPCR(tagPair)`: Retrieves or registers a new PCR reaction indexed by tag pair (case-insensitive).
- `NGSLibrary.Markers`: Map of primer pairs → `Marker` objects.
  - Lazy initialization via `GetMarker()` for new primers.

### 5. **Validation & Consistency Checks**
- `CheckTagLength()`: Ensures all registered tags have uniform length per direction.
- `CheckPrimerUnicity()`: Validates no primer is reused across markers; prevents self-complementary pairs.

### 6. **Batch Processing & Parallelism**
- `ExtractBarcodeSlice(sequences, options)`: Processes a slice of reads.
  - Configurable via `Options` (fluent API):
    - Mismatch/indel budgets.
    - Error handling (`discardErrors`, `OptionUnidentified`).
    - Parallel workers, batch size.
- `ExtractBarcodeSliceWorker()`: Returns a reusable worker for concurrent pipelines.

### 7. **Distance Metrics**
- `Hamming(s1, s2)`: Counts mismatches between equal-length strings.
- `Levenshtein(s1, s2)`: Computes edit distance (supports indels).

### 8. **Sample Identification**
- `TagExtractor`: Extracts forward/reverse tags from primer-flanked regions.
- `SampleIdentifier`:
  - Matches extracted tags to known samples using configured strategy (`strict`, `hamming`, or `indel`).
  - Returns best-matching sample, distance, and proposed tags.

---

## Design Highlights
- **Memory-efficient**: Uses reference-counted sequences (`Recycle()`).
- **Error-aware**: Rich error propagation (stored in `DemultiplexMatch.Error` or annotations).
- **Flexible tag design**: Supports fixed, delimited (exact), and indel-resilient tags.
- **Extensible via options**: Functional setters for clean, testable configuration.

---

## Use Case
Ideal for **metabarcoding or targeted amplicon sequencing**, where samples are multiplexed using unique dual barcodes. Ensures high specificity (unique tag pairs) and sensitivity (error-tolerant matching).
