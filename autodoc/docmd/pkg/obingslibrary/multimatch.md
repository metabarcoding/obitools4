# Semantic Description of `obingslibrary` Package

The `obingslibrary` package provides core functionality for **multiplexed high-throughput sequencing (HTS) data processing**, specifically designed to extract, validate, and assign biological samples from NGS reads using **dual-indexed barcodes** flanked by primers.

## Key Functionalities

1. **Primer & Tag Matching Structures**  
   - `PrimerMatch`: Encodes location, orientation (`Forward`), mismatch count, and marker identity of primer hits.
   - `TagMatcher`: Functional interface for extracting sample-specific tags from sequence regions.

2. **Distance Metrics**  
   - `Hamming`: Counts character mismatches between equal-length strings (for strict mismatch tolerance).
   - `Levenshtein`: Computes edit distance allowing insertions/deletions (for indel-tolerant matching).

3. **Tag Extraction Strategies**  
   - `lookForTag`: Extracts delimited tags (e.g., between two identical delimiters).
   - `lookForRescueTag`: Robustly extracts tags despite indels or variable delimiter lengths.
   - `*Fixed/Delimited/RescueTagExtractor` methods: Support three tag formats per primer direction (fixed-length, delimited with exact delimiters, or rescue-tolerant).

4. **Marker & Library Abstraction**  
   - `NGSLibrary`: Holds a map of primer pairs (`PrimerPair`) to `Marker` objects.
   - Each `Marker`: Defines forward/reverse primer sequences, tag specifications (length/spacer/delimiter/indels), and sample-to-tag mappings.

5. **Tag Assignment & Sample Identification**  
   - `TagExtractor`: Extracts forward/reverse tags from primer-flanked regions and annotates them.
   - `SampleIdentifier`: Matches extracted tags to known samples using configurable matching modes:
     - `"strict"`: Exact match only.
     - `"hamming"`: Closest tag by Hamming distance (substitutions).
     - `"indel"`: Closest tag by Levenshtein distance.
   - Annotates results with matching mode, distances, and proposed tags.

6. **Multi-Barcode Extraction**  
   - `ExtractMultiBarcode`: Scans a full sequence for primer pairs (forward/reverse + their complements), detects valid amplicon intervals, and:
     - Extracts the internal barcode region.
     - Assigns tags → sample via `SampleIdentifier`.
     - Annotates each barcode with primer matches, errors, directionality.
   - Handles both orientations (`forward` and `reverse`) of the amplicon.

7. **Parallel Processing Integration**  
   - `ExtractMultiBarcodeSliceWorker`: Returns a reusable worker function for batch processing sequences, supporting options like indel tolerance and mismatch limits.

## Use Case  
This package enables **demultiplexing** of NGS reads in amplicon-based workflows (e.g., metabarcoding), where samples are labeled with unique dual barcodes. It ensures robustness against sequencing errors and supports flexible tag design.
