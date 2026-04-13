# `obilua` Package: Biosequence Lua Bindings

The `obilua` Go package provides **Lua bindings** for biological sequence objects (`obiseq.BioSequence`) used in the OBITools4 ecosystem. It enables scripting and automation of sequence analysis directly from Lua.

## Core Functionality

- **Type Registration**: Registers a new userdata type `BioSequence` in the Lua state, exposing methods and constructors.
- **Constructor**:  
  ```lua
  BioSequence.new(id, sequence[, definition]) →BioSequence```
- **Accessors & Mutators**:
  - `id()`, `sequence()`, `definition()` – get/set identifiers and sequence data.
  - `qualities([table])` – handle PHRED-quality scores (as Lua table or string).
  - `count()`, `taxid()` – numeric abundance and taxonomic ID.
- **Taxonomy Integration**:
  - `taxon([Taxon])` – get/set taxonomic assignment via integrated taxonomy engine.
- **Attributes**:
  - `attribute(name[, value])` – arbitrary metadata storage (supports tables, strings, numbers).
- **Sequence Operations**:
  - `len()` – length of the sequence.
  - `has_sequence()`, `has_qualities()` – boolean checks for presence of data.
- **Computation & Transformation**:
  - `subsequence(start, end)` – extract a region.
  - `reverse_complement()` → BioSequence.
  - `md5()`, `md5_string()` – compute sequence checksums (raw bytes or hex string).
- **Serialization**:
  - `fasta([format])`, `fastq([format])` – output in FASTA/FASTQ, supporting `"json"` or `"obi"` header formats.
  - `string([format])` – smart formatting: FASTQ if qualities present, else FASTA.

## Implementation Notes

- Uses `gopher-lua` for interpreter integration.
- UserData wrapping ensures type safety and GC management of Go-backed objects.
- Error handling via Lua `ArgError` or `RaiseError`.
