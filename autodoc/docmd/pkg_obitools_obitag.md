# `obitag` — Geometric & Alignment-Based Taxonomic Assignment Module

The **OBITools4 `obitag`** package enables high-throughput taxonomic assignment of biological sequences using two complementary strategies:  
- A **geometric approach** based on landmark-based coordinate mapping and distance minimization.  
- An **alignment-aware heuristic** combining k‑mer pre-screening with LCS-based similarity scoring.  

Both modes integrate seamlessly into OBITools4 pipelines, support parallelization via `obiiter`, and enrich output sequences with rich metadata for downstream analysis.

---

## Public Functionalities

### 1. Reference Database Handling  
- **`CLIRefDB()`**: Loads a reference database from file (FASTA/FASTQ/OBI/etc.) into `BioSequenceSlice`.  
- **`CLISaveRefetenceDB()`**: Persists the loaded/processed reference DB to disk, with optional compression and parallel I/O.

### 2. CLI Configuration & Options  
- **`TagOptionSet()`**: Defines command-line flags:  
  - `-R/--reference-db`: *Required* input reference file.  
  - `--save-db`: Optional path to save processed DB (supports `.gz`, parallel write).  
  - `-G/--geometric`: Enables *experimental* geometric mode (faster, approximate).  
- **`CLIGeometricMode()`**, `CLIRefDBName()`, `CLIRunExact()`: Runtime accessors for internal state.

### 3. Geometric Taxonomic Assignment  
- **`ExtractLandmarkSeqs()`**: Retrieves reference sequences annotated with non-default `LandmarkID`s, ordered by ID.  
- **`ExtractTaxonSet()`**: Maps each landmark sequence to its taxonomic node (panics on missing taxa).  
- **`MapOnLandmarkSequences()`**: Computes a coordinate vector for any query sequence:  
  - Measures LCS distances to each landmark → yields point in *landmark space*.  
- **`FindGeomClosest()`**: Finds reference sequences with minimal Euclidean distance in landmark space;  
  - Resolves ties using LCS-based sequence identity (higher = better).  
- **`GeomIdentify()`**: Assigns taxonomy to a query:  
  - If best identity >50% → LCA of matching references’ taxa, weighted by geometric distance.  
  - Else → assigns root taxon (`taxid=1`).  

### 4. Alignment-Based Taxonomic Assignment  
- **`MatchDistanceIndex()`**: Maps a distance value to the closest taxon in `distanceIdx`:  
  - Binary search on sorted keys; falls back to root if no match.  
- **`FindClosests()`**: Retrieves top matching references for a query:  
  - Pre-screening via **4-mer overlap** (`Common4Mer`).  
  - Refinement using LCS alignment scoring.  
  - Returns: top matches, edit distance (`maxe`), sequence identity (%), best match ID & indices.  
- **`Identify()`**: Full taxonomic classification:  
  - Uses `FindClosests()`, precomputed reference indices (`OBITagRefIndex`), and LCA over matches.  
  - Assigns root taxon if no confident match; populates metadata (see below).  

### 5. Pipeline & Worker Integration  
- **`GeomIdentifySeqWorker()`** / `IdentifySeqWorker()`: Wraps assignment logic into reusable sequence workers.  
- **`CLIGeomAssignTaxonomy()`** / `CLIAssignTaxonomy()`: High-level CLI entry points:  
  - Filters/validates references, builds indexes (4-mer + taxon).  
  - Launches parallel batch processing via `obiiter`.  

---

## Output Metadata (Added to Assigned Sequences)

| Attribute | Description |
|-----------|-------------|
| `"scientific_name"` | Taxonomic name of assigned node. |
| `"obitag_rank"`     | Rank (e.g., `species`, `genus`). |
| `"obitag_bestid"`   | Sequence identity (%) of best match. |
| `"obitat_min_dist"` | Minimal geometric distance (landmark space). |
| `"obitag_match_count"` | Number of matching references used for LCA. |
| `"obitat_coord"`    | Landmark coordinates (geometric mode only). |
| `"obitation_similarity_method": "geometric"` | or `"alignment"`. |

---

## Design Principles

- **Dual-mode flexibility**: Choose between speed (`--geometric`) or accuracy (default alignment).  
- **LCS-centric robustness**: Avoids full alignments; uses longest common subsequence for noise-tolerant scoring.  
- **Index reuse**: Caches taxonomic indexes per reference to avoid recomputation in batch mode.  
- **Fail-safe fallbacks**: Missing taxa or low identity → root taxon (`taxid=1`).  
- **Scalability**: Parallel workers, batched iteration (`IBatchOver`), and optional compression.

---

## Dependencies

- `obitools4/obiiter`, `obiseq`, `obitax`, ` obialign`, `obikmer`  
- Standard I/O via `obiconvert`, `obiformats`

