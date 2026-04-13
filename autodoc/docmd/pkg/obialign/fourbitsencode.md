# Semantic Description of `obialign` Package

The `obialign` package provides low-level utilities for efficient nucleotide sequence encoding and decoding, specifically designed for bioinformatics alignment tasks.

- **Core functionality**: Encodes IUPAC nucleotide symbols (including ambiguous codes like `R`, `Y`, `N`) into compact 4-bit binary representations.
- **Binary encoding scheme**: Each bit in a byte corresponds to one canonical nucleotide: A (bit 0), C (bit 1), G (bit 2), T (bit 3).  
- **Ambiguity support**: Codes like `R` (A/G) set both corresponding bits (`0b0101`). Fully ambiguous `N` sets all four bits (`0b1111`).
- **Gap/missing handling**: Symbols `.` and `-`, as well as non-nucleotide characters, map to `0b0000`.
- **Memory efficiency**: The encoding avoids allocations via optional buffer reuse.
- **Lookup tables**:
  - `_FourBitsBaseCode`: Maps ASCII nucleotide characters (lowercased via `nuc & 31`) to their binary code.
  - `_FourBitsBaseDecode`: Inverse mapping for human-readable output (not exported, used internally).
- **Integration**: Works with `obiseq.BioSequence`, a generic biological sequence container from the OBITools4 ecosystem.

The `Encode4bits` function enables fast, space-efficient sequence processing—ideal for high-throughput sequencing data where alignment speed and memory usage are critical.
