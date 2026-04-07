# `obitagpcr` Package: Semantic Feature Overview

The `obitagpcr` package extends the OBITools4 ecosystem with CLI-ready, high-performance tools for **tag-based amplicon sequencing data processing**, focusing on consistent read orientation and robust sample demultiplexing using molecular barcodes.

## Core CLI Configuration

- **`TagPCROptionSet()`**  
  Adds a `--reorientate` flag to the CLI parser. When enabled, ensures all reads are stored in a *forward-strand orientation* relative to the expected PCR primers—by reverse-complementing reads originally aligned in the opposite direction.

- **`OptionSet()`**  
  Aggregates all required option sets for tag-PCR workflows:  
  - `obipairing.OptionSet()` — controls paired-end read assembly (e.g., overlap, identity thresholds),  
  - `obimultiplex.MultiplexOptionSet()` — enables sample demultiplexing via barcode matching,  
  - `TagPCROptionSet()` — injects the reorientation behavior.  

- **`CLIReorientate(cli)`**  
  Returns a boolean indicating whether reorientation is enabled, allowing downstream components to conditionally apply strand correction.

## Sequence Processing Pipeline

- **Paired-end assembly**  
  Uses `obipairing.AssemblePESequences()` to merge forward/reverse reads into consensus amplicons, respecting user-defined parameters:  
  - `minOverlap`, `minIdentity` — alignment stringency,  
  - gap/penalty parameters (`gapOpen`, `scale`) for accurate overlap resolution.

- **Barcode extraction & validation**  
  Applies a compiled NGS filter (`CLINGSFilter`) to extract and validate barcodes from consensus sequences. Only reads with *exactly one* valid barcode (no error flags) proceed to demultiplexing.

- **Sample assignment & metadata annotation**  
  Successful matches assign:  
  - `forward_tag`, `reverse_barcode` — raw barcode sequences,  
  - `obimultiplex_direction` — strand orientation relative to primer set (e.g., `"F"`, `"R"`),  
  - `obimultiplex_mismatches` — number of barcode mismatches,  
  - sample name (`obimultiplex_sample`) and experiment ID.  

  Annotations are propagated to *both* reads in the original pair.

- **Reorientation logic**  
  When `--reorientate` is active: reads assigned in reverse orientation (`"R"`) are reversed-complemented *before* final output, ensuring all consensus amplicons share a uniform forward orientation—critical for downstream alignment or variant calling.

- **Error handling & filtering**  
  Failed demultiplexing (e.g., no match, ambiguous barcode) flags reads with `obimultiplex_error`. By default:  
  - Unidentified reads are discarded, *or*  
  - Saved to a dedicated file via `CLIUnidentifiedFileName(cli)`.

- **Parallelization & scalability**  
  Leverages goroutines and batched iterators (`obidefault.ParallelWorkers()`) to maximize throughput across CPU cores.

- **Observability**  
  Optional statistics tracking (`withStats`) and structured logging (e.g., `"Worker started"`, `"Barcode filter passed"`), aiding debugging and performance profiling.

## Integration & Use Cases

Designed for **amplicon/metabarcoding workflows** where:  
- PCR amplifies both DNA strands, leading to bidirectional reads;  
- Primer positions are fixed and known (enabling orientation-aware assembly);  
- Consistent strand direction improves accuracy in alignment, clustering, or taxonomic assignment.  

Built on core OBITools4 modules (`obiseq`, `obiiter`, `obialign`, `obimultiplex`), it integrates cleanly into modular NGS pipelines while preserving modularity and CLI extensibility.
