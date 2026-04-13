# `obik match`: Semantic Description of the Subcommand

The `runMatch` function implements the **k-mer-based sequence matching** subcommand (`obik match`) of OBITools4. It enables rapid identification of query sequences (e.g., reads) against one or more reference k-mer indexes stored in `.kdi` files.

### Core Functionality

1. **Index Loading**  
   Opens a k-mer index (`KmerSetGroup`) from disk, retrieving metadata (k-mer size `k`, minhash dimensionality `m`, number of partitions, and sets). Supports matching against specific reference sets via glob-like patterns or all available sets.

2. **Sequence Input**  
   Reads input biological sequences (FASTA/FASTQ) using the standard `obiconvert` reader, preserving paired-end information.

3. **Parallel Query Preparation**  
   Sequences are split across multiple goroutines (`nworkers`). Each worker:
   - Extracts a batch of sequences.
   - Preprocesses them into *prepared queries* via `ksg.PrepareQueries`, which computes k-mers and hashes them for efficient lookup.

4. **Batch Accumulation & Query Merging**  
   Prepared queries from workers are merged incrementally in a single consumer goroutine:
   - Batches and their sequences are accumulated.
   - Queries are merged using `obikmer.MergeQueries`, updating sequence indices to reflect the combined batch.
   - When accumulated k-mer count reaches `defaultMatchQueryThreshold` (10M), the merged work is flushed to the matching stage.

5. **Batch Matching & Annotation**  
   For each accumulated batch and selected reference set:
   - `ksg.MatchBatch` performs all-vs-all k-mer matching, returning positions where matches occur.
   - Results are attached to original sequences as attributes (e.g., `kmer_matched_ref_genome: [12, 45]`).
   - Annotated batches are forwarded to the output stream.

6. **Output Streaming**  
   Matched sequences (now annotated) are written to stdout or a file via `CLIWriteBioSequences`, respecting paired-end structure.

### Key Design Principles

- **Zero shared mutable state** between pipeline stages.
- **Memory efficiency**: Large query sets are processed in chunks (threshold-based flushing).
- **Parallelism at multiple levels**:
  - Query preparation across CPUs.
  - Internal parallelization of `MatchBatch` per partition (handled by the k-mer engine).
  - Single-threaded accumulation to avoid race conditions on merged queries.
  
### Output Semantics

Each output sequence carries one or more attributes indicating which reference sets it matched, and at what positions — enabling downstream filtering, profiling (e.g., taxonomic assignment), or visualization.
