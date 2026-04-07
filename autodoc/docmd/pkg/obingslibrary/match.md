# Demultiplexing Functionality in `obingslibrary`

This package provides tools for **demultiplexing NGS reads** by matching them against known primer pairs and extracting associated barcodes.

## Core Types

- `DemultiplexMatch`: Struct holding alignment results for forward/reverse primers, mismatches, barcode coordinates (`BarcodeStart`, `BarcodeEnd`), and metadata (e.g., sample/experiment info via `PCR`). Includes error handling.

## Key Methods

- **`Match(sequence)`**:  
  Scans the input `BioSequence` against all primer pairs in `NGSLibrary.Markers`. Returns a populated `DemultiplexMatch` if any primer pair matches.

- **`ExtractBarcode(sequence, inplace)`**:  
  Uses the result of `Match()` to:
    - Extract the barcode region (if valid: non-dimer).
    - Reverse-complement if read is in reverse orientation (`IsDirect == false`).
    - Annotate the sequence with:
      - Primer names and match details (positions, mismatches).
      - Direction (`direct`/`reverse`).
      - Sample/experiment info (if assignment succeeds), or error message.

## Behavior Notes

- **Primer dimer detection**: If `BarcodeStart > BarcodeEnd`, the read is flagged as a primer dimer and not extracted.
- **Error handling**: Errors (e.g., no match, sample unassignment) are stored in `match.Error` and propagated as annotations.
- **Annotation richness**: Output sequences carry rich metadata (sample, experiment, primers, errors), supporting downstream filtering/analysis.

## Dependencies

- Uses `logrus` for fatal logging (e.g., subsequence extraction failure).
- Integrates with `obiseq.BioSequence` for sequence representation and manipulation.
