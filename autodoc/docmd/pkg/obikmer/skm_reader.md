# SKM File Reader for Super-Kmers

This Go package provides a binary file reader (`SkmReader`) for `.skm` files, which store *super-kmers* — compact representations of DNA sequences using 2-bit encoding.

## Core Functionality

- **Binary Format Parsing**: Reads structured data from `.skm` files, where each record contains:
  - A 2-byte little-endian integer specifying the sequence length.
  - Packed nucleotide data, where every byte encodes up to four bases (2 bits per base).

- **Decoding Logic**: Converts packed 2-bit codes (`00`, `01`, `10`, `11`) to nucleotide characters using the mapping:  
  `{ 'a', 'c', 'g', 't' }`.

- **Memory-Efficient Reading**: Uses buffered I/O (64 KiB buffer) for fast sequential access.

- **Streaming Interface**: `Next()` returns the next super-kmer as a struct with:
  - `Sequence`: decoded nucleotide byte slice.
  - `Start`, `End`: positional metadata (currently fixed to full length).

- **Resource Management**: Provides a clean `.Close()` method for file handle cleanup.

## Use Case

Designed for high-performance processing of large genomic datasets (e.g., in k-mer analysis or sequence indexing), where storage size and read speed are critical.
