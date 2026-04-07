# `obilua` Package: Lua Bindings for BioSequence Slicing

This Go module provides **Lua scripting support** for biological sequence manipulation via the `obilua` package. It exposes a custom Lua type, `"BioSequenceSlice"`, wrapping Go’s `*obiseq.BioSequenceSlice` to enable high-level sequence operations in Lua.

## Core Features

- **Type Registration**: Registers `BioSequenceSlice` as a userdata type in Lua with metatable support.
- **Constructor**: `new([capacity])` creates a new slice (optionally pre-sized).
- **Indexing & Assignment**: `slice[i] = seq` or `seq = slice[i]`, with bounds checking.
- **Dynamic Operations**:  
  - `push(seq)`: Append a sequence.  
  - `pop()`: Remove and return the last sequence.
- **Length Query**: `len()` returns number of sequences in slice.

## Output Formatting

Provides multiple export methods to format all contained sequences:

- `fasta([format])`: Returns FASTA string (supports `"json"` or `"obi"` headers).
- `fastq([format])`: Returns FASTQ string (same format options as above).
- `string([format])`: Smart formatter:  
  - Uses FASTQ if *all* sequences have quality scores.  
  - Falls back to FASTA otherwise.

## Design Notes

- All methods validate input types and indices.
- Format selection is optional; defaults to `"obi"` header style unless specified as `"json"`.
- Integrates with `obiseq.BioSequence` and formatting utilities from the OBItools4 ecosystem.

This enables Lua users to process NGS data (e.g., FASTA/FASTQ) interactively within pipelines, leveraging Go’s performance and Lua’s expressiveness.
