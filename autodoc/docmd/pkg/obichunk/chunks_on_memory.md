# `ISequenceChunkOnMemory` Function — Semantic Description

The function `Isequencechunkonmemory`, from the Go package `obichunk`, implements **asynchronous in-memory chunking** of biological sequence data.

It consumes an iterator over `BioSequence` objects and distributes them into **heterogeneous batches** using a provided classifier. The core purpose is to group sequences by classification (e.g., sample, taxon, or feature), store each group in memory as a slice (`BioSequenceSlice`), and emit them sequentially via an output iterator.

Key features:
- **Parallel processing**: Each classification group (referred to as a *flux*) is processed in its own goroutine.
- **Thread-safe aggregation**: A mutex ensures safe concurrent updates to shared `chunks` and `sources` maps.
- **Lazy emission**: Batches are emitted only after all classification groups have been fully processed (`jobDone.Wait()`).
- **Ordered output**: Batches are emitted in increasing `order` index (0, 1, …), preserving determinism despite parallel internal processing.
- **Error handling**: Critical failures (e.g., channel retrieval errors) terminate the program with `log.Fatalf`.

Input:
- An iterator (`obiiter.IBioSequence`) of raw sequences.
- A `*obiseq.BioSequenceClassifier`, used to route each sequence into a classification bucket.

Output:
- A new iterator yielding `BioSequenceBatch` objects, each containing all sequences belonging to one classification group and its source identifier.

Use case: Efficient parallel preprocessing of high-throughput sequencing data into sample- or taxon-specific batches for downstream analysis.
