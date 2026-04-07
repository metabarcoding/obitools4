## Semantic Description of `obirefidx` Package

The `obirefidx` package implements a taxonomic indexing pipeline for biological sequences, enabling efficient reference-based classification via alignment-free and alignment-based methods.

### Core Functionality

- **`IndexSequence(seqidx, references, kmers, taxa, taxo)`**  
  Computes a *taxonomic signature* for a query sequence by comparing it against reference sequences. It:
  - Identifies Least Common Ancestors (LCAs) between the query and all references using a cached LCA lookup.
  - Groups reference sequences by their shared LCAs with the query across taxonomic ranks.
  - Uses **4-mer common counts** for fast pre-filtering of candidates.
  - Performs local alignment (via `FastLCSScore` or exact distance `D1Or0`) to compute error counts (substitutions + indels).
  - Builds a strictly increasing vector `closest[]` of minimal alignment errors per taxonomic rank.
  - Maps each error threshold to the most specific matching taxon (`"Taxon@Rank"`), stored in a map keyed by error count.

- **`IndexReferenceDB(iterator)`**  
  Processes an entire reference database:
  - Loads sequences and filters out those lacking valid taxonomic IDs.
  - Precomputes **4-mer frequency tables** for all sequences to accelerate k-mer comparisons.
  - Parallelizes indexing in batches (10 seqs/worker), using `IndexSequence` per sequence.
  - Attaches the resulting taxonomic index (`obitag`) to each *copy* of the sequence via `SetOBITagRefIndex`.
  - Returns an iterator over batches, optionally displaying a progress bar.

### Key Technical Features

- **Taxonomy-aware filtering**: Exploits hierarchical taxonomic structure to limit alignment scope.
- **Hybrid similarity search**: Combines *k*-mer sharing (fast) with LCS-based alignment (accurate).
- **Caching & optimization**: LCA results are cached; memory for alignments is reused via a shared `matrix`.
- **Parallelization**: Uses goroutines and channels to process sequences concurrently.
- **Robust error handling & logging**: Leverages `logrus` for detailed diagnostics and progress tracking.

### Output Format

Each indexed sequence carries a map `map[int]string`, where:
- Keys = alignment error counts (e.g., mismatches + gaps),
- Values = taxonomic labels like `"Homo@genus"` or `"Vertebrata@subphylum"`,
enabling rank-specific classification thresholds.
