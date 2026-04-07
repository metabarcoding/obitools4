# Semantic Description of `obiformats` Package Functionalities

The `obiformats` package provides robust, streaming-aware chunking utilities for processing large biological sequence files (e.g., FASTA/FASTQ) in a memory-efficient and parallel-friendly manner.

- **`PieceOfChunk`**: A rope-like linked buffer structure enabling efficient concatenation and partial reading of large data streams without full materialization. Supports dynamic chaining (`NewPieceOfChunk`, `Next()`) and final packing into a contiguous slice via `Pack()`.

- **`FileChunk`**: Encapsulates one chunk of raw data (`*bytes.Buffer`) or its rope representation, tagged with source file name and positional order for ordered downstream processing.

- **`ChannelFileChunk`**: A typed channel (`chan FileChunk`) enabling concurrent, pipeline-style data ingestion—ideal for parallel parsing or streaming workflows.

- **`LastSeqRecord`**: A callback type (`func([]byte) int`) used to locate the end of a complete biological record (e.g., last newline after full FASTQ entry), ensuring chunks split only at valid boundaries.

- **`ReadFileChunk()`**: Core function that:
  - Reads from an `io.Reader` in configurable chunks (`fileChunkSize`);
  - Uses a probe string (e.g., `"@M0"` for FASTQ) to early-exit non-matching segments and avoid unnecessary parsing;
  - Extends chunks incrementally (e.g., +1 MB) until a full record boundary is found via `splitter`;
  - Returns data as an ordered stream of `FileChunk`s on a channel, closing it upon EOF;
  - Optionally packs rope buffers to contiguous memory (`pack` flag), balancing speed vs. RAM usage.

- **Key semantics**:  
  - *Chunking by record integrity*, not fixed byte size — prevents splitting biological entries.  
  - *Lazy evaluation*: only reads ahead when needed to find record boundaries.  
  - *Streaming-first design* — supports large files without full loading into memory.

This package is foundational for scalable, robust parsing of high-throughput sequencing data in the OBITools4 ecosystem.
