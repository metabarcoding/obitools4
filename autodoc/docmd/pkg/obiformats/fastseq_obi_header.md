# OBIFormats Package: Semantic Description

The `obiformats` package provides parsing and formatting utilities for **OBI-compliant FASTA headers**, enabling structured annotation of biological sequences.

- It supports parsing key-value annotations embedded in sequence definitions (e.g., `key=value;`), including nested dictionaries.
- Three core parsing functions detect value types:  
  - `__match__key__`: Identifies assignment patterns (`Key = ...`).  
  - `__obi_header_value_numeric_pattern__`: Matches floats/integers (e.g., `42.0;`).  
  - `__obi_header_value_string_pattern__`: Matches quoted strings (e.g., `'example';`).  
  - `__match__dict__`: Parses balanced `{...}` blocks, handling nested structures and string delimiters.

- Boolean detection (`__is_true__/__is_false__`) handles multiple case variants (e.g., `true`, `True`, `TRUE`).

- The main entry point, **`ParseOBIFeatures(text string, annotations obiseq.Annotation)`,**  
  iteratively extracts key-value pairs from a header string and populates an `Annotation` map.  
  - Numeric values are stored as integers if they have no fractional part.  
  - Dictionary-like strings (e.g., `{'a':1,'b':2}`) are JSON-unmarshalled into typed maps:  
    - `*_count` → `map[string]int`,  
    - `merged_*` → wrapped in a statistics object (`obiseq.StatsOnValues`).  
    - `*_status`/`*_mutation` → `map[string]string`.  

- **`ParseFastSeqOBIHeader(sequence *obiseq.BioSequence)`** applies parsing to a sequence’s definition line, moving annotations into its metadata map and preserving leftover text.

- **`WriteFastSeqOBIHeade(buffer *bytes.Buffer, sequence)`** serializes annotations back into OBI header format:  
  - Strings and booleans use `key=value;`.  
  - Maps/dicts are JSON-encoded, then single-quoted for compatibility.  
  - Special handling ensures `obiseq.StatsOnValues` are safely marshalled.

- **`FormatFastSeqOBIHeader(sequence)`** returns the formatted header as a string (zero-copy via `unsafe.String` for performance).

- Designed to interoperate with the broader OBITools4 ecosystem (`obiseq`, `obiutils`), supporting both human-readable and machine-processable sequence metadata.
