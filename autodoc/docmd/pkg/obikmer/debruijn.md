# Semantic Description of the `obikmer` Package

This Go package implements a **De Bruijn graph** for efficient k-mer manipulation and sequence assembly, primarily used in bioinformatics (e.g., metagenomic read error correction or consensus building).

### Core Functionalities

- **K-mer Encoding**: K-mers are encoded as `uint64` using 2 bits per nucleotide (A=0, C=1, G=2, T=3), supporting IUPAC ambiguity codes via the `iupac` map.
- **Reverse Complement Handling**: The `revcompnuc` table enables nucleotide-wise reverse complementation.
- **Graph Construction**: The `DeBruijnGraph` struct maintains a map from k-mer hashes to integer weights (e.g., observed counts), with helper masks for bit manipulation (`kmermask`, `prevc/g/t`).

### Graph Operations

- **Node Queries**:  
  - `Previouses()` / `Nexts()`: Return predecessor/successor k-mers in the graph.  
  - `MaxNext()` / `MaxHead()`: Find neighbors or heads (sources) with maximum weight.
- **Path Exploration**:  
  - `MaxPath()`: Greedily traces the highest-weight path from a head.  
  - `LongestPath()`: Explores all heads to find the path with maximum cumulative weight (optionally bounded in length).  
  - `HaviestPath()`: Uses Dijkstra-like priority queue to find the *heaviest* (sum-weight) path, with cycle detection via DFS (`HasCycle()`).

### Consensus & Filtering

- **Consensus Generation**:  
  - `BestConsensus()` returns a sequence from the greedy max-weight path.  
  - `LongestConsensus(id, min_cov)` trims low-coverage ends using a coverage threshold (mode-based).
- **Weight Statistics**:  
  - `MaxWeight()`, `WeightMean()`, `WeightMode()` provide distribution summaries.  
  - `FilterMinWeight(min)` removes low-count nodes.
- **Decoding**:  
  - `DecodeNode()` converts a k-mer index to its DNA string.  
  - `DecodePath()` reconstructs the full consensus from a path.

### I/O & Diagnostics

- **GML Export**: `WriteGml()` outputs a directed graph in Graph Modelling Language (for visualization), with edge thickness and labels reflecting weights.
- **Hamming Distance**: `HammingDistance()` computes edit distance between two encoded k-mers using bit operations.
- **Sequence Insertion**: `Push()` adds a biosequence (with count weight) to the graph, expanding all IUPAC variants recursively.

### Dependencies & Design

- Leverages `obiseq` for sequence representation and `logrus`/`slices`/`heap` from Go’s stdlib.
- Designed for scalability and speed, using bit-level operations to minimize memory footprint.

Overall: a robust k-mer graph engine for *de novo* assembly, error correction, and consensus recovery in high-throughput sequencing data.
