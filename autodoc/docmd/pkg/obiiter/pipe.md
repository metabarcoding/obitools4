# Semantic Description of `obiiter` Package Features

This Go package provides functional-style utilities for processing biological sequence data (e.g., FASTQ/FASTA), modeled via the `IBioSequence` interface.

- **`Pipeable`**: A function type representing a unary transformation on an `IBioSequence`.  
- **`Pipeline(start, parts...)`**: Composes a sequence of `Pipeable` operations into a single executable pipeline. It applies transformations sequentially: input → start → part₁ → … → output.

- **`(IBioSequence).Pipe(start, parts...)`**: A convenience method enabling fluent chaining of transformations directly on a sequence object.

- **`Teeable`**: A function type for operations that split input into two independent output streams (e.g., filtering + logging).

- **`(IBioSequence).CopyTee()`**: A high-level tee operation that duplicates the input stream into two identical, concurrently readable `IBioSequence` instances.  
  - Uses goroutines to ensure non-blocking parallel consumption.
  - Ensures proper lifecycle management: closing the second stream when the first is closed.  
  - Preserves paired-end status (`MarkAsPaired`) if applicable.

Together, these features support modular, composable, and concurrent biosequence processing pipelines—ideal for scalable NGS data workflows.
