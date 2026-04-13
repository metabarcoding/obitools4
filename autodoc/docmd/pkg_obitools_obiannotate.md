# `obiannotate`: Semantic Description of Public Features

The `obiannotate` package delivers modular, composable sequence annotation workers for biological sequences (FASTA/FASTQ) within the OBITools4 ecosystem. Each worker returns an `obiseq.SeqWorker`, enabling declarative pipeline construction via chaining or conditional execution. All functionality is exposed through both programmatic and CLI interfaces.

## 1️⃣ Attribute Management  
Workers manipulate sequence annotations (metadata slots) with fine-grained control:
- **`DeleteAttributesWorker(keys)`**: Removes specified annotation keys; silently skips missing ones.  
- **`ToBeKeptAttributesWorker(keys)`**: Retains only listed keys; discards all others.  
- **`ClearAllAttributesWorker()`**: Strips *all* annotations from each sequence.  
- **`RenameAttributeWorker(mapping)`**: Renames keys using a dict (e.g., `{"old": "new"}`); skips records if source key is absent.

## 2️⃣ Sequence Editing  
Direct manipulation of sequence content and derived metadata:
- **`CutSequenceWorker(start, end)`**: Extracts subsequence from `start` to `end` (1-based; supports negative indices). Fails with error or discards sequence on invalid bounds.  
- **`AddSeqLengthWorker()`**: Adds `seq_length = len(sequence)` annotation.  
- **`EvalAttributeWorker(expr, target_slot=None)`**: Evaluates Python expressions (e.g., `"seq_length > 200"`) to set annotations; used internally by `EditAttributeWorker`.

## 3️⃣ Taxonomic Annotation  
Enriches sequences with taxonomic context using NCBI taxonomy:
- **`AddTaxonAtRankWorker(rank)`**: Adds taxon name at specified rank (e.g., `"species"`) to slot `taxon_at_rank`.  
- **`AddTaxonRankWorker()`**: Infers and annotates taxonomic rank (e.g., `"species"`).  
- **`AddScientificNameWorker()`**: Adds `scientific_name = "Homo sapiens"`-style label.  
- **`AddTaxonomicPathWorker()`**: Adds full lineage path (semicolon-separated).  

## 4️⃣ Pattern Matching  
Detects DNA motifs with tolerance for mismatches/indels:
- **`MatchPatternWorker(pattern, max_errors=0, allow_indel=False)`**:  
  - Scans both strands via reverse-complement.  
  - Annotates: `slot_location` (start/end), `slot_match`, and `slot_error`.  
  - Uses **Aho-Corasick** for efficient multi-pattern search (file-based via `obicorazick.AhoCorasickWorker`).

## 5️⃣ CLI-Driven Pipeline Construction  
Bridges command-line flags to composable workers:
- **`CLIAnnotationWorker(args)`**: Builds a composite worker from CLI flags (e.g., `--pattern`, `--taxonomic-rank`).  
- **`CLIAnnotationPipeline(args)`**: Wraps the worker in a conditional pipeline (using `obigrep` predicates) and parallelizes via multiprocessing.

## 6️⃣ Utility & Validation  
- **`CLIHasPattern(pattern)`**: Returns a worker that filters sequences matching `pattern`.  
- **`CLICut(start, end)`**: Returns a cut worker for CLI usage.  
- All workers validate inputs (e.g., malformed `--cut` triggers fatal exit with log).

All public features are **stateless**, composable via `ChainWorkers`, and designed for high-throughput, scriptable annotation workflows.
