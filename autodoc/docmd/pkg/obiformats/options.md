# Semantic Description of `obiformats` Package Functionalities  

The `go` package `obiformats` provides a flexible, configuration-driven framework for handling biological sequence data (e.g., FASTA/FASTQ) and associated metadata. Its core component is the `Options` type, which encapsulates user-defined settings via an immutable configuration pattern using functional setters (`WithOption`).  

Key capabilities include:  
- **I/O control**: file handling options (e.g., `OptionCloseFile`, `OptionsAppendFile`), compression support (`OptionsCompressed`), and batch processing modes (e.g., `FullFileBatch`, custom `BatchSize`).  
- **Parallelism & performance tuning**: configurable number of workers (`OptionsParallelWorkers`) and memory buffer size (via `TotalSeqSize`).  
- **Sequence parsing/formatting**: pluggable header parsers/writers for FASTA/FASTQ (e.g., `OptionsFastSeqHeaderParser`, `OptionFastSeqDoNotParseHeader`), with support for quality scores (`OptionsReadQualities`).  
- **CSV export**: granular control over columns (ID, sequence, quality, taxon, count), separators (`CSVSeparator`), NA values (`CSVNAValue`), and auto-inferred keys (`CSVAutoColumn`).  
- **Taxonomic metadata integration**: toggles for taxid, scientific name, rank, path (with/without root), parent relationships (`OptionsWithTaxid`, `OptionWithoutRootPath`), and U→T conversion for ambiguous bases.  
- **Advanced features**: feature table inclusion (`WithFeatureTable`), pattern matching support (`OptionsWithPattern`), and paired-end read handling via `WritePairedReadsTo`.  
- **Metadata extensibility**: arbitrary metadata fields can be attached via `OptionsWithMetadata`, with automatic cleanup (e.g., removal of `"query"` when pattern mode is active).  

All options are initialized with sensible defaults (e.g., `batch_size`, `parallel_workers`) and can be composed using the `MakeOptions` constructor. This design enables declarative, reusable configuration across sequence processing pipelines in OBITools4.
