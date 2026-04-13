# Aho-Corasick-Based Sequence Analysis in `obicorazick`

This Go package provides efficient pattern-matching utilities for biological sequence data, leveraging the Aho-Corasick algorithm.

## Core Components

- **`AhoCorazickWorker(slot string, patterns []string) obiseq.SeqWorker`**  
  Builds *multiple* Aho-Corasick matchers in parallel (batched to manage memory), then returns a `SeqWorker` function.  
  - Scans each sequence *forward* and its reverse complement.
  - Counts total matches (`slot`), forward-only (`_Fwd`) and reverse-complement-specific (`_Rev`) matches.
  - Attaches match counts as sequence attributes.

- **`AhoCorazickPredicate(minMatches int, patterns []string) obiseq.SequencePredicate`**  
  Compiles a *single* matcher and returns a predicate function.  
  - Returns `true` if the number of matches ≥ `minMatches`.
  - Useful for filtering sequences (e.g., taxonomic assignment or contamination detection).

## Technical Highlights

- **Batched compilation**: Large pattern sets are split into chunks (default `10⁷` patterns/batch) to avoid memory overload.
- **Parallelization**: Matcher construction uses goroutines, scaled by `obidefault.ParallelWorkers()`.
- **Progress tracking**: Optional CLI progress bar via `progressbar/v3`, enabled globally.
- **Logging & debugging**: Uses Logrus for info/debug messages; logs match counts per sequence.

## Use Cases

- Rapid screening of sequences against large reference databases (e.g., primers, barcodes, contaminants).
- Filtering or annotating sequences based on pattern presence/abundance.
