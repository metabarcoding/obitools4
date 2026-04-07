# FASTQ Parsing Module (`obiformats`)

This Go package provides robust, streaming-capable parsing of FASTQ files — a standard format for storing nucleotide sequences along with quality scores.

## Core Functionalities

- **`EndOfLastFastqEntry(buffer []byte) int`**  
  Locates the start position (`@`) of the last complete FASTQ entry in a byte buffer using state-machine scanning from end to beginning. Returns `-1` if no valid entry found.

- **`FastqChunkParser(...)`**  
  Returns a parser function for processing FASTQ data from an `io.Reader`. Handles:
  - Header parsing (`@id [definition]`)
  - Sequence normalization (uppercase → lowercase, `U→T` conversion if enabled)
  - Quality score shifting (`quality_shift`)
  - Strict validation (e.g., `+` line, matching sequence/length)

- **`FastqChunkParserRope(...)`**  
  Optimized parser for rope-based input (`PieceOfChunk`), avoiding unnecessary memory copies. Uses direct line-by-line scanning.

- **Batched File Parsing (`_ParseFastqFile`, `ReadFastq`, etc.)**  
  Enables concurrent, chunked parsing of large files:
  - Splits input into chunks using `ReadFileChunk`
  - Uses configurable parallel workers (`nworker`)
  - Pushes parsed batches to an iterator interface

- **Convenience I/O Wrappers**
  - `ReadFastqFromFile(filename, ...)`: Parses a file by name.
  - `ReadFastqFromStdin(...)`: Reads FASTQ from standard input.

## Key Options & Features

- **Quality handling**: Optional quality extraction (`with_quality`), configurable offset (`quality_shift`)
- **Uracil-to-Thymine conversion**: `UtoT` flag for RNA→DNA normalization
- **Header annotation parsing**: Optional post-parsing header interpretation via `ParseFastSeqHeader`
- **Batch sorting & full-file mode**: Supports both streaming and complete-file aggregation

## Design Highlights

- **Memory-efficient chunking** with overlap-aware boundary detection (`EndOfLastFastqEntry`)
- **Strict error reporting**: Fails fast on malformed FASTQ (e.g., invalid chars, length mismatch)
- **Integration with `obiseq`, `obiiter`**: Returns typed biological sequence slices and iterator streams compatible with the broader OBITools4 ecosystem.
