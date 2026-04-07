# `obicleandb` Package Overview

The `obicleandb` package delivers semantic curation and trust scoring for biological sequences (e.g., DNA barcodes) within the OBITools4 ecosystem. It combines taxonomic consistency checks, alignment-based discrimination tests, and statistical confidence estimation to ensure high-fidelity sequence datasets for downstream analysis.

## Core Functionalities

### 1. **Input & Taxonomy Integration**
- Loads reference taxonomies (e.g., NCBI) via `obioptions.LoadTaxonomyOptionSet`.
- Parses heterogeneous inputs (FASTA/FASTQ) using `obiconvert.InputOptionSet`, supporting streaming and format auto-detection.
- Integrates taxonomic lineage information into sequence metadata for downstream filtering.

### 2. **Taxonomy-Guided Dereplication & Filtering**
- `ICleanDB` orchestrates a pipeline that first filters sequences by required taxonomic ranks (e.g., species, genus).
- Dereplicates identical sequences *within* taxonomic groups (e.g., collapse duplicates per `taxid`), preserving only one representative per unique sequence–taxon pair.
- Ensures minimal taxonomic resolution before scoring (e.g., requires at least genus-level assignment).

### 3. **Sequence Trust Scoring**
- `SequenceTrust`: Computes *local* confidence as  
  \[
    s = 1 - \frac{1}{n + 1}
  \]  
  where `n` is the count of identical sequences sharing taxonomic labels—interpreting duplicates as empirical validation.
- `SequenceTrustSlice`: Computes *global* confidence via pairwise alignment distances (LCSS scores) among group members.  
  - Normalizes observed intra-group distance by the median pairwise distance across all groups (`obicleandb_median`).  
  - Estimates effective sample size (`obicleandb_trusted_on`) using group composition and redundancy.

### 4. **Higher-Rank Discrimination (Mann–Whitney U Test)**
- `MakeSequenceFamilyGenusWorker` tests whether a sequence’s alignment scores to conspecifics are significantly better than to outgroups at genus/family level.
- Uses `obialign.FastLCSScore` for rapid approximate alignment scoring on grouped sequences.
- Outputs a *p*-value stored in `obicleandb_trusted`, indicating confidence that the sequence belongs to its assigned higher-rank taxon.

### 5. **Efficient Distance Storage**
- `diagCoord` implements compact triangular indexing for pairwise distance matrices, reducing memory footprint by ~50% while enabling fast lookup.

### 6. **Pipeline Orchestration**
- `ICleanDB` unifies all steps: input → taxonomy loading → filtering/dereplication → trust scoring.
- Returns an iterator of cleaned, annotated sequences with standardized attributes.

## Output Attributes

| Attribute | Description |
|----------|-------------|
| `obicleandb_trusted` | Final confidence score (probability of correct taxonomic assignment) |
| `obicleandb_trusted_on` | Effective group size used for scoring (accounts for redundancy) |
| `obicleandb_level` | Taxonomic rank used in discrimination test (`genus`, `family`, or `"none"`) |
| `obicleandb_median` | Median pairwise LCSS distance used as normalization baseline |

## Design Principles

- **Modularity**: Workers (e.g., `SequenceTrust`, `MakeSequenceFamilyGenusWorker`) are composable and reusable.
- **Parallelism**: Batched processing via `obidefault` settings for scalability across large datasets.
- **Robustness**: Gracefully handles sparse taxonomy, small group sizes, and missing labels.

This package enables rigorous pre-processing of metabarcoding datasets—critical for reducing false positives in OTU/ASV inference and ecological interpretation.
