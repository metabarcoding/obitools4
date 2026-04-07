# Semantic Description of `obiannotate` Package

The `obiannotate` package provides a suite of sequence annotation workers for processing biological sequences (e.g., FASTA/FASTQ) in the OBITools4 ecosystem. Each function returns an `obiseq.SeqWorker`, enabling functional composition and pipeline integration.

- **Attribute Management**:  
  - `DeleteAttributesWorker`: removes specified annotation keys.  
  - `ToBeKeptAttributesWorker`: retains only user-specified attributes (others deleted).  
  - `ClearAllAttributesWorker`: strips all annotations.  
  - `RenameAttributeWorker`: renames annotation keys via a mapping dictionary.

- **Sequence Editing**:  
  - `CutSequenceWorker`: extracts subsequence between positions (supports negative indexing); returns error or discards sequence on failure.  
  - `EvalAttributeWorker`: dynamically sets annotation fields using expression strings (via chaining with `EditAttributeWorker`).  
  - `AddSeqLengthWorker`: adds a `"seq_length"` annotation.

- **Taxonomic Annotation**:  
  - `AddTaxonAtRankWorker`: annotates taxon at specified ranks (e.g., `"species"`).  
  - `AddTaxonRankWorker`: infers and sets taxonomic rank.  
  - `AddScientificNameWorker`: adds scientific name annotation.

- **Pattern Matching**:  
  - `MatchPatternWorker`: detects user-defined DNA patterns (with error tolerance, indels allowed optionally), annotating match location (`slot_location`), sequence (`slot_match`), and errors (`slot_error`). Supports both strands via reverse-complement search.

- **CLI Integration**:  
  - `CLIAnnotationWorker`: constructs a composite worker based on command-line flags (e.g., pattern matching, taxonomic annotation, attribute filtering).  
  - `CLIAnnotationPipeline`: wraps the worker in a conditional pipeline (using selection predicates from `obigrep`) and parallelizes execution.

- **Advanced Matching**:  
  - Uses Aho-Corasick automata (`obicorazick.AhoCorasickWorker`) for efficient multi-pattern matching.

All workers are composable via `ChainWorkers`, enabling modular, declarative annotation pipelines for high-throughput sequence processing.
