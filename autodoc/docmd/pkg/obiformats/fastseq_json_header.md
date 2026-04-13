This Go package `obiformats` provides semantic parsing and serialization utilities for FASTQ/FASTA sequence headers encoded in JSON format, primarily used within the OBITools4 framework.

- **JSON Parsing Helpers**:  
  It defines internal functions (`_parse_json_map_*`, `_parse_json_array_*`) to convert JSON objects/arrays into typed Go maps and slices (`map[string]string`, `[]int`, etc.), using the high-performance [`jsonparser`](https://github.com/buger/jsonparser) library for streaming parsing.

- **Header Interpretation**:  
  `_parse_json_header_` interprets a FASTQ/FASTA header string containing embedded JSON metadata. It extracts and assigns:
  - Core fields (`id`, `definition`, `count`)
  - Specialized OBITools annotations (e.g., `"obiclean_weight"`, `"taxid"` with optional taxonomic ranks)
  - Generic annotations of any JSON type (string, number, bool, array, object), preserving numeric precision where possible.

- **Sequence Annotation Enrichment**:  
  `ParseFastSeqJsonHeader` parses the header of a `BioSequence`, extracting JSON metadata into its annotations map and reconstructing non-JSON text as the new definition.

- **Serialization Support**:  
  `WriteFastSeqJsonHeader` and `FormatFastSeqJsonHeader` serialize sequence annotations back into JSON format, appending them to a buffer or returning as string — enabling round-trip compatibility for annotated sequences.

- **Error Handling**:  
  Uses `log.Fatalf` on parsing failures, ensuring malformed headers fail fast during processing.

In summary: *structured JSON header ↔ BioSequence annotation mapping*, optimized for metabarcoding workflows.
