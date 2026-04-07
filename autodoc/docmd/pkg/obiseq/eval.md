# Semantic Description of `obiseq` Expression-Based Workers

This module provides **expression-driven transformation workers** for biological sequence objects (`BioSequence`). It leverages a custom expression language (via `OBILang`) to dynamically compute values based on sequence metadata and content.

## Core Components

- **`Expression(expression string)`**:  
  Returns a function that evaluates the given expression in context. The evaluation scope includes:
  - `annotations`: sequence annotations (metadata).
  - `sequence`: the full `BioSequence` object itself.

- **`EditIdWorker(expression string)`**:  
  A sequence worker that updates the *ID* of a `BioSequence` by evaluating the expression.  
  - On success: sets `sequence.Id()` to string representation of result.
  - On failure: logs and returns an error with context.

- **`EditAttributeWorker(key string, expression string)`**:  
  A sequence worker that sets a *custom attribute* (identified by `key`) on the sequence, using evaluated expression result.  
  - Supports arbitrary metadata enrichment.
  - Errors are reported with sequence ID and failed expression.

## Use Cases

- Generate new IDs from annotation fields (e.g., `"gene_" + annotations["locus_tag"]`).
- Compute and store derived attributes (e.g., GC content, ORF length) as sequence metadata.
- Apply conditional logic or transformations across large sets of sequences in pipelines.

All workers conform to the `SeqWorker` interface, enabling composition and chaining.
