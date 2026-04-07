# `MakeSetAttributeWorker` Functionality Overview

The function `MakeSetAttributeWorker(rank string) obiiter.SeqWorker` constructs a reusable sequence-processing worker for taxonomic annotation.

- **Input validation**: It first verifies that the provided `rank` is part of a predefined taxonomic hierarchy (`taxonomy.RankList()`). If invalid, it terminates execution with an informative error.

- **Worker construction**: It returns a closure (`obiiter.SeqWorker`) — essentially a function that transforms biological sequences.

- **Core behavior**: For each input `*obiseq.BioSequence`, it calls `taxonomy.SetTaxonAtRank(sequence, rank)`. This likely assigns or updates the taxonomic label (e.g., species, genus) at the specified rank in the sequence’s metadata.

- **Purpose**: Enables modular, pipeline-friendly taxonomic annotation — e.g., in bioinformatics workflows where sequences must be annotated hierarchically (e.g., from phylum down to species).

- **Design pattern**: Follows the *functional factory* and *worker interface* patterns, promoting composability in sequence processing pipelines.

- **Side effects**: Modifies the input `BioSequence` *in-place* (via mutation of its taxonomic metadata), then returns it.

- **Use case example**:  
  ```go
  worker := MakeSetAttributeWorker("species")
  seq = worker(seq) // annotates `seq` with species-level taxon
  ```

- **Assumptions**:  
   - `taxonomy.SetTaxonAtRank` exists and handles rank-specific taxon assignment.  
   - Taxonomic ranks are ordered, finite, and validated (e.g., `["domain", "phylum", ..., "species"]`).  
   - Sequences carry mutable taxonomic metadata.

- **Error handling**: Fails fast on invalid rank input, preventing silent misannotation.
