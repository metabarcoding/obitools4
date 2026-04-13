Here's a **structured, semantic description** (≤200 lines) of the public API provided by the `obiseq` Go package, written in English and Markdown format:

```markdown
# BioSequence Attribute & Sequence Management (`obiseq`) — Public API Overview

The `obiseq` package provides a high-performance, thread-safe framework for representing and manipulating biological sequences (DNA/RNA/protein) in Go. It supports rich metadata, annotations, quality scores, taxonomic integration, and efficient batch processing—ideal for NGS pipelines like OBITools4.

## Core Sequence Representation

- `BioSequence`: Immutable-like container for sequence data (`[]byte`), ID, definition, qualities, features, and annotations.
- `NewBioSequence(...)`, `NewEmptyBioSequence(cap)`: Constructors supporting initialization with ID, sequence, definition, and optional qualities.
- `Id()`, `Definition()`: Accessors for core metadata fields (ID normalized to lowercase).
- `Sequence()` / `String()`: Returns the sequence as a copy or human-readable string.
- `Len()`, `HasSequence()` / `Composition()`: Length, presence check, and nucleotide composition (`a,c,g,t,o`).
- `MD5()`, `MemorySize()` / `Recycle()`: Integrity checksum, memory footprint estimation, and safe object pooling reset.

## Attribute & Annotation System

- `Annotations()`, `HasAnnotation(key)`: Read-only access to generic metadata map.
- Thread-safe via internal mutex (`AnnotationsLock()`).
- `GetAttribute(key)`, `SetAttribute(key, value)` / typed getters (`GetIntAttribute(...)`) with automatic type coercion.
- `Keys()` & `HasAttribute(key)`: Enumerate and check presence of attributes (including `"id"`, `"sequence"`).
- `AttributeKeys(skip_map, skip_definition)`: Aggregates all attribute keys across a collection.

## Quality & Feature Support

- `Qualities()` / `SetQualities(...)`: Per-base quality scores (Phred+40 default).
- `HasQualities()`, `Write(...)`, `Clear()` / quality ASCII conversion.
- `Features()`: Optional raw feature table (e.g., GenBank/EMBL).

## Pairing & Taxonomy

- `PairTo(p)`, `IsPaired()`, `UnPair()` / batch pairing for read-pairs.
- Taxonomic annotation:  
  - `Taxid()`, `SetTaxid(...)`, `Taxon(taxonomy)`  
  - Rank-specific: `SetSpecies()`, `SetGenus()` / generic via `SetTaxonAtRank(rank)`  
  - Full path & LCA: `Path()`, `SetTaxonomicDistribution(...)`

## Classification, Filtering & Transformation

- Classifiers:  
  - `AnnotationClassifier`, `DualAnnotationClassifier` / predicate-based (`PredicateClassifier`)  
  - Hashing, rotation & composite strategies (e.g., `CompositeClassifier`)
- Predicates:  
  - Length, abundance (`IsMoreAbundantOrEqualTo`) / regex matching on ID/sequence  
  - Expression-based (`ExpressionPredicat`), paired-end support
- Workers:  
  - `EditIdWorker`, `EditAttributeWorker` (via OBILang expressions)  
  - Taxonomic annotators (`MakeSetSpeciesWorker`, `LCA`) / reverse-complement & subsequence workers

## Collection Management & Efficiency

- `BioSequenceSlice`: Optimized batch container with:
  - Pool-aware allocation (`NewBioSequenceSlice`, `EnsureCapacity`)  
  - Efficient push/pop, sorting (on count/length), and merging
- `Merge(...)`: Sequence & slice-level consensus with stat propagation.
- Slice/annotation pooling:  
  - `GetSlice`, `RecycleSlice` / annotation recycling via pools
- Iterators:  
  - `Kmers(k)`: Lazy k-mer generator using Go’s new iterator protocol.

## Utility & Extension

- IUPAC support: `SameIUPACNuc(a, b)` for ambiguity-aware base comparison.
- Reverse complement: `ReverseComplement(inplace)`, mutation coordinate adjustment (`_revcmpMutation`).
- Subsequence extraction: `Subsequence(from, to, circular)` with quality & annotation preservation.
- Expression extensions (via OBILang):  
  - `gc`, `gcskew` / `elementof`, `sprintf`, `ifelse`

All methods ensure correctness via safe type conversions, locking semantics, and graceful fallbacks—enabling scalable bioinformatics workflows.
