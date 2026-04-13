# Semantic Description of `obichunk.ISequenceSubChunk`

The function `ISequenceSubChunk` in the `obichunk` package implements **parallel, class-based sorting and batching of biological sequences**, preserving input order within each batch while reordering across batches by classification code.

## Core Functionality

- **Input**:  
  - An iterator over `BioSequence` batches (`obiiter.IBioSequence`)  
  - A sequence classifier (`obiseq.BioSequenceClassifier`) assigning each sequence a numeric class code  
  - A number of worker goroutines (`nworkers`), defaulting to system-configured parallelism  

- **Processing**:  
  - Each worker consumes its own iterator split and classifier clone, enabling concurrent batch processing.  
  - For each incoming `BioSequenceBatch`:  
    - If the batch has >1 sequence: sequences are extracted, classified into `code`, and sorted *in-place* by class code.  
    - Consecutive sequences with the same `code` are grouped into new batches; a new batch is emitted upon code change.  
    - If the batch has ≤1 sequence, it’s passed through unchanged (but reordered with a new order ID).  

- **Ordering Mechanism**:  
  - Uses `atomic.AddInt32` to assign strictly increasing order IDs (`nextOrder`) across workers, preserving deterministic inter-batch ordering.  
  - Sorting within batches is performed via a custom `sort.Interface` implementation using closures for flexible comparison logic (here, by ascending class code).  

- **Output**:  
  - Returns a new iterator (`obiiter.IBioSequence`) emitting batches grouped by classification code, with globally ordered batch IDs.  
  - Workers are coordinated via `newIter.Done()`/`Wait()/Close()`, ensuring clean termination.

## Semantic Purpose

Enables efficient, parallel **grouping of sequences by taxonomic or functional class** (e.g., OTU assignment), optimizing downstream processing that requires sorted/class-ordered input — e.g., consensus building, alignment, or read merging per group.
