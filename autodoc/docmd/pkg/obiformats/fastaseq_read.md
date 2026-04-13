# FASTA Parser Module (`obiformats`)

This Go package provides robust, streaming-capable parsing of FASTA-formatted nucleotide sequences. It supports both standard and rope-based (memory-efficient) input handling.

## Core Functionalities

- **`FastaChunkParser(UtoT bool)`**  
  Returns a parser function for in-memory byte streams. Converts `U→T` if enabled (for RNA/DNA normalization). Validates headers, identifiers, and sequences; rejects invalid characters or malformed entries.

- **`FastaChunkParserRope(...)`**  
  Parses FASTA directly from a `PieceOfChunk` rope structure, avoiding full data materialization. Optimized for large files.

- **`ReadFasta(reader io.Reader, ...)`**  
  High-level API to parse FASTA from any `io.Reader`. Uses chunked reading with parallel workers (configurable via options). Supports full-file batching and header annotation parsing.

- **`ReadFastaFromFile(...)` / `ReadFastaFromStdin(...)`**  
  Convenience wrappers for file and stdin inputs, including source naming and empty-file handling.

- **`EndOfLastFastaEntry(...)`**  
  Helper to locate the last complete FASTA entry in a buffer, enabling safe chunked streaming without splitting records.

## Key Features

- **Strict validation**: Ensures entries start with `>`, contain valid identifiers, and only use allowed sequence characters (`a-z`, `- . [ ]`).
- **Case normalization**: Converts uppercase to lowercase; optional `U→T` conversion.
- **Whitespace handling**: Ignores spaces/tabs in sequences, preserves line breaks only for parsing structure.
- **Parallel processing**: Configurable worker count via options; batches results by source and order for downstream sorting/aggregation.
- **Integration with `obiseq`/`obiiter`**: Yields typed sequence objects (`BioSequence`) and batched iterators compatible with OBITools4 pipelines.

## Design Highlights

- Minimal allocations via rope-based parsing (`extractFastaSeq`).
- Graceful error reporting with context (source, identifier, invalid char position).
- Extensible via `WithOption` pattern for header parsing and batching behavior.
