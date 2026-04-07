# `obirefidx` Package: Semantic Overview

The `obirefidx` package implements a **taxonomic reference indexing pipeline** for high-throughput sequencing data, optimized for family-level classification. It combines *k*-mer-based pre-filtering with alignment-aware similarity scoring to build compact, taxonomically annotated reference indexes—enabling fast and accurate read assignment in metabarcoding workflows.

---

## Public Functionalities

### 1. **Reference Database Indexing Pipeline**

#### `IndexSequence(seqidx int, references []obiio.BioSeq, kmers obikmer.Table4mer, taxa map[string]TaxonID, taxo TaxonomySlice) (map[int]string)`
Computes a **taxonomic error-profile** for one query sequence against all references:
- Uses cached LCA lookups to group references by shared taxonomic ancestors.
- Filters candidate sets using 4-mer overlap counts (fast).
- Performs local alignment (`FastLCSScore` or `D1Or0`) to compute substitution+indel error counts.
- Builds a strictly increasing vector of minimal errors per taxonomic rank (e.g., genus, family).
- Outputs a map: `error_count → "Taxon@Rank"` (e.g., `{0: "Homo@genus", 3: "Primates@order"}`).

> ✅ *Key insight*: Taxonomic resolution degrades predictably with alignment error.

---

#### `IndexReferenceDB(iter obiio.SequenceIterator) (obiio.BatchedSequenceIterator)`
Processes an entire reference database into indexed batches:
- Validates sequences: skips those without valid taxonomic IDs.
- Precomputes 4-mer frequency tables for all sequences (via `obikmer.Table4mer`).
- Parallelizes indexing over chunks of 10 sequences using worker goroutines.
- Calls `IndexSequence` for each sequence and attaches the result (`obitag`) to a copy.
- Returns an iterator over batches, optionally displaying progress.

> ✅ *Design note*: Memory reuse and batched I/O ensure scalability to large databases.

---

### 2. **Clustering & Deduplication**

#### `MakeStartClusterSliceWorker(chunkSize int, identityThreshold float64) (func([]obiio.BioSeq) []ClusterSlice)`
Performs **greedy hierarchical clustering** at family-level identity (hardcoded ≥90%):
- Uses LCSS alignment with error tolerance derived from `identityThreshold`.
- For each sequence, outputs:
  - `clusterid`: ID of its cluster centroid (head).
  - `clusterhead`: boolean flag indicating if it *is* the head.
  - `clusteridentity`: alignment-based identity to the centroid.

> ✅ *Use*: Reduces redundancy before indexing—only centroids are re-indexed for efficiency.

---

### 3. **Taxonomy & Geography-Aware Indexing**

#### `GeomIndexSesquence(seqidx int, references []obiio.BioSeq, taxa map[string]TaxonID, taxo TaxonomySlice) (map[int]string)`
Computes a **spatially-aware taxonomic index**:
- Retrieves geographic coordinates (lat/long) of the query sequence; fails if missing.
- Computes Euclidean squared distances to all others in parallel.
- Sorts neighbors by distance while preserving original indices (`obiutils.Order`).
- Iteratively updates the LCA between query and neighbors, recording:
  - `distance → "Taxon@Rank"` map.
- Stops early upon reaching root taxonomy.

> ✅ *Use case*: Models taxonomic uncertainty bands based on nearest neighbors’ location + taxonomy.

---

### 4. **Worker Utilities & Taxonomy Annotation**

#### `MakeSetFamilyTaxaWorker()`, `MakeSetGenusTaxaWorker()`, etc.
Helper workers to annotate sequences with family/genus/species taxonomy:
- Uses `Taxonomy.LCA()` and cached taxon IDs to assign ranks.
- Parallelized over sequence batches (10 seqs/worker).
- Ensures all indexed sequences carry full taxonomic context.

---

### 5. **CLI Integration**

#### `OptionSet(options *getoptions.GetOpt)`
Configures CLI options for the `obiuniq` tool:
- Delegates to `obiconvert.OptionSet(false)` (no verbose logging).
- Enables only options relevant for reference deduplication.
- Ensures consistent, minimal interface across OBITools4 tools.

---

## Technical Highlights

| Feature | Description |
|--------|-------------|
| **Parallelization** | Goroutines with `obidefault.ParallelWorkers()` for indexing, distance computation & clustering. |
| **Memory Efficiency** | Reused buffers (`matrix`), batched processing, and sequence deduplication reduce RAM footprint. |
| **Caching** | LCA lookups, 4-mer tables, and alignment matrices are cached to avoid recomputation. |
| **Logging & Validation** | Structured logging via `logrus`; panics on critical errors (e.g., missing taxonomy). |
| **Progress Tracking** | Optional progress bar via `progressbar/v3` during large DB processing. |

---

## Output Format

Indexed sequences carry a map:  
```go
map[int]string // error_count → "Taxon@Rank"
```
Example:
```json
{
  0: "Homo@genus",
  2: "Hominoidea@superfamily",
  5: "Primates@order"
}
```
Enables **rank-specific classification thresholds** (e.g., “assign to genus if ≤2 errors”).

---

## Use Cases

- **Metabarcoding classification**: Rapid assignment of reads to reference families.
- **Reference curation**: Cluster & deduplicate large databases before indexing.
- **Ecological inference**: Estimate taxonomic uncertainty from spatial proximity + taxonomy.

> 📌 *Design principle*: Align with OBITools4’s philosophy—modular, parallelizable, and taxonomically aware.
