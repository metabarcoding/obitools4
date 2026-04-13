# `obisplit` Package: Semantic Description

The `obisplit` package enables **targeted splitting of biological sequences** using user-defined pattern pairs (e.g., primers, barcodes), supporting approximate matching and robust annotation of resulting fragments—ideal for demultiplexing in metabarcoding or amplicon sequencing pipelines.

## Core Concepts

- **`SplitSequence`**: Represents a pattern pair (forward/reverse) with an associated group name. Used to define searchable molecular tags.
- **`Pattern_match`**: Encapsulates a detected pattern instance, including name, genomic coordinates (1-based), error count, and orientation.

## Pattern Detection (`LocatePatterns`)

Scans a sequence for all forward/reverse pattern occurrences using **fuzzy matching** (mismatches and optionally indels):

- Accepts raw or indexed sequences for efficient lookup.
- Detects matches with configurable error tolerance (default: ≤4 mismatches).
- Normalizes coordinates and reverse-complements backward-strand matches.
- Deduplicates overlapping hits by retaining the match with fewer errors.

## Sequence Splitting (`SplitPattern`)

Divides input sequences into fragments **between matched pattern pairs**, producing annotated output:

- Each fragment is labeled with:
  - `obisplit_frg`: Fragment number (1-based).
  - `obisplit_nfrg`: Total fragment count.
  - `obisplit_group`: Pattern-pair name (e.g., `"primerA-primerB"`), or `"extremity"` for terminal regions.
  - `obisplit_set`: Relevant pattern set (e.g., `"primerA"`), or `"NA"`.
  - `obisplit_location`: Genomic span (1-based, inclusive).
- Includes left/right pattern metadata: name, matched substring, and error count.

## Pipeline Integration

- **`SplitPatternWorker`**: Wraps splitting logic as a reusable `SeqWorker`, compatible with OBITools4’s streaming infrastructure.
- **`CLISplitPipeline`**: CLI entry point integrating pattern detection and splitting into a parallelizable, configurable pipeline.

## Configuration & Usage

- **CSV-based config**: Maps `tag` sequences to `pcr_pool` identifiers (required columns: `tag`, optionally `reverse_tag`).
- **CLI flags**:
  - `-C, --config`: Load pattern definitions from CSV.
  - `--template`: Output sample config for rapid setup.
  - `--pattern-error N`: Max mismatches allowed (default: 4).
  - `--allows-indels`: Enable insertion/deletion-aware matching.
- **Error handling**: Validates config structure, pattern compilation, and file access; logs fatal issues.

## Design Goals

Optimized for **high-throughput amplicon processing**, `obisplit` bridges pattern detection and fragment extraction with minimal assumptions—ensuring flexibility for diverse molecular tagging schemes.
