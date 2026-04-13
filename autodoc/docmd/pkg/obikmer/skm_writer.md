# `.skm` File Format and `SkmWriter` Functionality

The Go package `obikmer` provides a binary writer for `.skm` (super-kmer) files, optimized for compact storage of DNA sequences.

- **Purpose**: Efficiently serialize *super-kmers* (long k-mers) into a binary format.
- **Format per super-kmer**:
  - `len: uint16 LE` — length of the sequence in bases (little-endian, 2 bytes).
  - `data: ⌈len/4⌉ bytes` — nucleotide sequence encoded as **2 bits per base**, packed tightly.

- **Encoding scheme**:
  - `A → 00`, `C → 01`, `G → 10`, `T → 11`.
  - Padding: trailing bits in the final byte are zeroed if `len % 4 ≠ 0`.

- **Implementation details**:
  - Uses buffered I/O (`bufio.Writer` with 64 KiB buffer) for performance.
  - `NewSkmWriter(path)` opens/creates the file and returns a writer instance.
  - `Write(sk SuperKmer)` encodes sequence length, then packs bases using a lookup (`__single_base_code__[seq[pos]&31]`).
  - `Close()` flushes buffers and closes the file handle.

- **Use case**: Ideal for high-throughput genomic preprocessing (e.g., indexing, sketching), where space and I/O speed matter.

- **Assumptions**: `SuperKmer` type exposes a `.Sequence []byte`; bases are ASCII (`A,C,G,T,a,c,g,t`) — `&31` normalizes to lowercase index.

- **Efficiency**: 4× compression vs. ASCII (1 byte/base → ~0.25 bytes/base), minimal overhead.
