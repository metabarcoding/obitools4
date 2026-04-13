# `ropeScanner` — Line-by-Line Text Scanning over a Rope Data Structure

The `obiformats` package provides the `ropeScanner`, an efficient line-oriented iterator over a *Rope* (a tree-based immutable string representation, implemented here as `PieceOfChunk`). This scanner supports streaming large texts without full materialization.

## Core Functionality

- **`newRopeScanner(rope *PieceOfChunk)`**  
  Constructs a new scanner starting at the root of the rope.

- **`ReadLine() []byte`**  
  Returns the next line (without trailing `\n`, or `\r\n`) as a byte slice.  
  - Returns `nil` when the end of the rope is reached.
  - Reuses internal buffers (`carry`) to handle lines spanning multiple nodes efficiently.
  - The returned slice aliases rope data and is only valid until the next call.

- **`skipToNewline()`**  
  Advances internal position to just after the next newline (`\n`), discarding content. Useful for skipping unwanted lines or headers.

## Implementation Highlights

- **Buffered carry-over**: Lines split across rope nodes are assembled incrementally in the `carry` buffer, which grows dynamically.
- **Cross-platform line endings**: Automatically strips `\r\n`, leaving only the content (no trailing CR).
- **Zero-copy where possible**: When a line fits entirely within one node and no carry exists, it returns a slice directly into the rope’s underlying data.

## Use Case

Ideal for parsing large text files or streams (e.g., OBIE/Obi formats) where memory efficiency and streaming behavior are critical—without loading the entire content into RAM.
