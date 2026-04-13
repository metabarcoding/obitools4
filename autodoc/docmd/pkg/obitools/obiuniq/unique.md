# Semantic Description of `CLIUnique` Functionality

The `CLIUnique` function implements a **dereplication pipeline** for biological sequence data (e.g., amplicons, reads), returning a deduplicated iterator of sequences (`obiiter.IBioSequence`).  

- **Core purpose**: Collapse identical or near-identical sequences while preserving metadata and counting occurrences.  
- **Input/Output**: Accepts a sequence iterator; outputs an iterator over unique sequences with abundance annotations.  
- **Chunking**: Processes data in configurable batches (`OptionBatchCount`) to manage memory and scalability.  
- **Sorting Strategy**: Supports in-memory or disk-based sorting via CLI flags (`--on-disk`), optimizing for large datasets.  
- **Singleton Handling**: Optionally filters out sequences observed only once (`--no-singleton`), configurable at runtime.  
- **Parallelization**: Leverages default parallel workers (`OptionsParallelWorkers`) to accelerate sorting/deduplication.  
- **Batching**: Uses default batch size (`OptionsBatchSize`) to balance throughput and memory usage.  
- **Missing Data**: Handles missing values (`OptionNAValue`) as defined by CLI arguments (e.g., `CLINAValue`).  
- **Statistics**: Enables optional per-category statistics collection (`OptionStatOn`) based on user-specified keys.  
- **Subcategorization**: Groups sequences by metadata keys (`OptionSubCategory`) to enable stratified dereplication (e.g., per sample, primer).  
- **Error Handling**: Logs fatal errors during pipeline initialization or execution using `log.Fatal`.  

The function integrates CLI-driven configuration into a modular, extensible chunk-based processing framework (`obichunk`), supporting both scalability and flexibility in high-throughput sequencing workflows.
