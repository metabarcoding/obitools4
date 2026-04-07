# `obitagpcr` Package: Paired-End Sequence Demultiplexing and Tagging

The `obitagpcr` package provides high-performance, parallelized demultiplexing and annotation of paired-end NGS reads using molecular barcodes (e.g., PCR tags). Its core function, `IPCRTagPESequencesBatch`, processes sequence pairs from an iterator and outputs annotated reads with sample-specific metadata.

## Key Functionalities

- **Paired-end assembly**: Reads are assembled into consensus sequences using `obipairing.AssemblePESequences`, with parameters for alignment gap, scale, overlap length (`minOverlap`), identity threshold (`minIdentity`), and fast alignment heuristics.

- **Barcode extraction**: A compiled NGS filter (`CLINGSFIlter`) extracts barcodes from each consensus. Only reads with a *single*, valid barcode (no error flags) are assigned to samples.

- **Metadata propagation**: Upon successful demultiplexing, barcode identity (`forward_tag`, `reverse_tag`), directionality (`obimultiplex_direction`), mismatches, sample name, and experiment ID are added as annotations to *both* reads in the pair.

- **Reorientation support**: If enabled (`CLIReorientate`), reverse-direction reads are reversed-complemented and re-paired to ensure consistent forward orientation of tags.

- **Error handling & filtering**: Unassigned reads (failed demultiplexing) are flagged with an `obimultiplex_error` annotation. By default, they can be discarded or saved to a separate file (`CLIUnidentifiedFileName`).

- **Parallel processing**: Uses goroutines and batched iteration to scale across CPU cores (`obidefault.ParallelWorkers()`), maximizing throughput.

- **Statistics & logging**: Optional stats collection (`withStats`) and structured log messages track pipeline stages (e.g., worker start/end, filtering decisions).

## Dependencies & Integration

Built on core `obitools4` modules (`obiiter`, `obiseq`, `obialign`, `obimultiplex`), it integrates seamlessly into larger NGS processing pipelines for metabarcoding and amplicon sequencing workflows.
