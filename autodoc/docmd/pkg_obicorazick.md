# `obicorazick`: Aho-Corasick-Based Sequence Analysis Package

`obicorazick` is a high-performance Go library for rapid pattern detection in biological sequences (e.g., FASTA/FASTQ), designed to scale efficiently with large pattern sets. Built on the Aho-Corasick algorithm, it enables concurrent scanning of sequences against thousands to millions of patterns—ideal for primer screening, contamination checks, or taxonomic classification.

## Public API

### `AhoCorazickWorker(slot string, patterns []string) obiseq.SeqWorker`

Constructs a **sequence worker function** that scans input sequences for matches against the provided `patterns`, using *multiple* Aho-Corasick automata compiled in parallel (batched internally to manage memory).  

- **Input**:  
  - `slot`: Name of the attribute field where match counts will be stored (e.g., `"primer_hits"`).  
  - `patterns`: List of DNA/RNA patterns (strings) to search for.  

- **Behavior**:  
  - Splits `patterns` into batches of ≤10⁷ items (configurable via environment).  
  - Compiles one Aho-Corasick matcher per batch in parallel (using `obidefault.ParallelWorkers()`).  
  - For each sequence: scans both the forward strand and its reverse complement.  
  - Records three counts as attributes on the sequence:  
    ```text
    <slot>         → total matches (forward + rev-comp)
    <slot>_Fwd     → forward-strand-only matches
    <slot>_Rev     → rev-comp-specific (i.e., not found on forward) matches
    ```  
  - Logs match counts at debug level (via Logrus).  

- **Use case**: Annotating sequences with pattern-hit statistics for downstream analysis (e.g., reporting primer coverage per read).

---

### `AhoCorazickPredicate(minMatches int, patterns []string) obiseq.SequencePredicate`

Returns a **boolean predicate function** that tests whether sequences contain ≥ `minMatches` occurrences of any pattern.  

- **Input**:  
  - `minMatches`: Minimum number of total matches required to pass the predicate.  
  - `patterns`: List of patterns (same format as above).  

- **Behavior**:  
  - Compiles a *single* Aho-Corasick matcher (no batching—assumes pattern set is moderate-sized or memory-safe).  
  - Scans only the forward strand (for efficiency in filtering contexts where rev-comp is unnecessary).  
  - Returns `true` if match count ≥ `minMatches`; otherwise `false`.  

- **Use case**: Filtering sequences—e.g., retain only reads containing ≥2 barcode primers, or discard those matching known contaminants.

---

## Implementation Notes (Non-Exported)

While not part of the public API, internal behavior includes:  
- **Batching logic**: Splits patterns to avoid memory exhaustion during automaton construction.  
- **Parallel compilation**: Uses goroutines + sync.WaitGroup, respecting `GOMAXPROCS`.  
- **Progress feedback**: Optional CLI progress bar (via `progressbar/v3`) when enabled globally.  
- **Logging**: Info/debug messages via Logrus (e.g., “Built 3 matchers in parallel” or “Sequence X: 5 total matches”).  

## Typical Workflows

1. **Annotation pipeline**:  
   ```go
   worker := AhoCorazickWorker("contam", contaminantDB)
   annotatedSeqs := obiseq.Map(worker, inputSequences)
   ```

2. **Filtering pipeline**:  
   ```go
   filter := AhoCorazickPredicate(1, barcodePatterns)
   filteredSeqs := obiseq.Filter(filter, inputSequences)
   ```

Designed for speed and memory efficiency in large-scale NGS data processing.
