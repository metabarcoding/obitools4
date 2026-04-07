# `FormatHeader` Function Type in `obiformats`

The `obiformats` package defines a core functional interface for sequence formatting within the OBITools4 ecosystem.

- **Package**: `obiformats`  
  Provides utilities for formatting biological sequences according to various output standards (e.g., FASTA, GenBank).

- **Type Definition**:  
  ```go
  type FormatHeader func(sequence *obiseq.BioSequence) string
  ```
  - A `FormatHeader` is a *function type* that takes a pointer to an `obiseq.BioSequence` and returns its formatted header as a string.

- **Semantic Role**:  
  Encapsulates the logic for generating *header lines* (e.g., `>id description`) in sequence file formats.  
  Decouples header formatting from core data structures (`BioSequence`), enabling modular and reusable format adapters.

- **Usage Context**:  
  - Used by writers/formatters to produce standardized headers when exporting sequences.  
  - Allows custom header generation (e.g., for MIxS-compliant metadata, user-defined tags).  
  - Supports polymorphism: different `FormatHeader` implementations can be swapped per output format.

- **Dependencies**:  
  - Relies on `obiseq.BioSequence`, the core sequence data model (ID, description, annotations, etc.).

- **Design Intent**:  
  Promotes clean separation of concerns: data (sequence) ↔ formatting logic.  
  Facilitates extensibility for new output formats without modifying core types.
