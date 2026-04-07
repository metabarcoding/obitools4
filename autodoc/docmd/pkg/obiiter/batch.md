# BioSequenceBatch: A Container for Ordered Biological Sequences

`BioSequenceBatch` is a structured data type encapsulating an ordered collection of biological sequences (`obiseq.BioSequenceSlice`) along with metadata: a `source` identifier and an integer `order`. It serves as a lightweight, immutable-friendly container for batch processing in bioinformatics pipelines.

## Core Properties
- **`source`**: String identifying the origin (e.g., file, pipeline stage).
- **`order`**: Integer defining processing sequence or priority.
- **`slice`**: Holds the actual sequences via `obiseq.BioSequenceSlice`.

## Key Functionalities
- **Construction**:  
  `MakeBioSequenceBatch(source, order, sequences)` creates a new batch.
- **Accessors**:  
  `Source()`, `Order()` return metadata; `Slice()` exposes the sequence slice.
- **Mutation (via copy)**:  
  `Reorder(newOrder)` returns a new batch with updated order.
- **Size & emptiness**:  
  `Len()` gives sequence count; `NotEmpty()` checks non-emptiness.
- **Consumption**:  
  `Pop0()` removes and returns the first sequence (FIFO behavior).
- **Safety**:  
  `IsNil()` detects uninitialized batches; a global `NilBioSequenceBatch` sentinel exists.

## Design Notes
- Instances are value types (struct), enabling safe copying.
- Operations follow Go idioms: methods return updated values rather than mutating in place (except internal slice mutation via `Pop0`).
- Designed for interoperability with the OBITools4 ecosystem (`obiseq` package).

This abstraction supports modular, traceable sequence processing workflows—ideal for pipeline stages where ordering and provenance matter.
