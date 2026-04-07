# `obikmersim`: K-mer–Based Sequence Similarity Analysis Package

`obikmersim` is a high-performance Python package for **k-mer–driven sequence comparison and alignment**, tailored for biological read analysis (e.g., amplicons, metagenomes). It enables rapid matching of query sequences against reference databases using efficient k-mer indexing, followed by localized alignment with quality-aware consensus refinement. Designed for scalability and flexibility, it supports sparse k-mer representations, orientation detection (forward/reverse-complement), and configurable filtering thresholds.

---

## Public API Overview

### 1. **K-mer Indexing & Matching Workers**
#### `MakeCountMatchWorker(reference_sequences, k=21, min_count=2, sparse=False)`
- **Purpose**: Build a `KmerMap` from reference sequences and match queries via shared k-mers.
- **Functionality**:
  - Indexes all *k*-mers (with optional sparsity mask) from reference sequences.
  - For each query, retrieves candidate references sharing ≥ `min_count` k-mers.
  - Returns annotated results: query ID, matched references, match count, *k*, and sparsity flag.
- **Use Case**: Fast pre-screening for taxonomic assignment or read clustering.

#### `MakeKmerAlignWorker(count_match_worker, delta=50, penalty_scale=1.0, gap_factor=-2)`
- **Purpose**: Perform *k*-mer–seeded local alignment with quality-aware consensus.
- **Functionality**:
  - Uses shared k-mers from `count_match_worker` to seed alignment candidates.
  - Runs local pairwise alignments (via internal aligner) and builds quality-weighted consensus (`ReadAlign`, `BuildQualityConsensus`).
  - Computes:
    - `% identity`
    - Residual similarity (k-mer–aware alignment score)
    - Alignment length & orientation (`+`/`−`)
  - Filters output by `min_identity=80%`, optional min alignment length.
- **Use Case**: Precise read assignment, error correction via consensus.

---

### 2. **CLI Configuration Options**
#### `KmerSimCountOptionSet`
- Defines CLI arguments for k-mer counting/matching:
  - `--kmer-size` (int, default=21)
  - `--sparse` (bool): Enable sparse k-mer masking
  - `--reference <file>`: Reference FASTA/FASTQ path(s)
  - `--min-count` (int, default=2): Minimum shared k-mer count
  - `--self`: Perform self-comparison (query = reference)

#### `KmerSimMatchOptionSet`
- Extends counting options with alignment scoring parameters:
  - `--delta` (int, default=50): Max k-mer separation for seeding
  - `--penalty-scale` (float, default=1.0): Mismatch/gap scaling factor
  - `--gap-factor` (float, default=−2): Gap penalty coefficient
  - `--fast-absolute`: Use fast absolute scoring (no dynamic programming)

#### Composite Sets
- `CountOptionSet` / `MatchOptionSet`: Combine k-mer options with generic I/O conversion settings (e.g., via `obiconvert`).

---

### 3. **CLI Helpers & Accessors**
#### `CLIKmerSize(args)`
- Returns parsed k-mer size from CLI args.

#### `CLIReference(args, format="fasta")`
- Loads reference sequences into memory (supports batched/parallel reading).

#### `CLISelf(args)`
- Returns boolean flag for self-comparison mode.

---

### 4. **Core CLI Wrappers**
#### `CLILookForSharedKmers(args)`
- Orchestrates k-mer counting/matching pipeline:
  - Builds `count_match_worker`
  - Iterates over query sequences (from stdin or file)
  - Outputs match annotations in structured format.

#### `CLIAlignSequences(args)`
- Runs full alignment pipeline:
  - Uses `count_match_worker` to seed candidates
  - Invokes `kmer_align_worker`
  - Outputs aligned pairs with identity, orientation, and quality metrics.

---

## Key Technical Features
- **Sparse K-mers**: Mask positions (e.g., Ns or degenerate bases) via bitmasks.
- **Orientation Handling**: Auto-detect reverse-complement matches during seeding/alignment.
- **Fast Heuristic Scoring**: Preliminary alignment score estimation before full path resolution (reduces compute).
- **Quality-Aware Consensus**: Integrates base quality scores during alignment refinement.
- **Configurable Filtering**: Thresholds on identity, length, and k-mer support.

---

## Typical Workflows
| Workflow | Tools Used |
|---------|------------|
| Taxonomic screening of amplicons | `CLILookForSharedKmers` + sparse mode |
| Read error correction via reference consensus | `CLIAlignSequences` with quality-aware alignment |
| *In silico* PCR specificity check | `CLISelf()` + min-count filtering |
| Large-scale metagenomic read assignment | Batched parallel execution with `CLIReference` |

---

## Output Format
Results are returned as structured records (e.g., dictionaries or dataclasses) with fields:
- `query_id`, `reference_ids`
- `match_count`, `kmer_size`, `sparse_mode`
- For alignments:  
  `%identity`, `alignment_length`, `orientation` (`+1`/`−1`)  
  `residual_similarity`, `consensus_quality`

All public functions are documented with type hints and include unit tests.
