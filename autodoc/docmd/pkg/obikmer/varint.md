# Variable-Length Integer Encoding/Decoding Utility

This Go package (`obikmer`) provides efficient serialization of `uint64` integers using **protobuf-style variable-length encoding (varint)**.

## Core Features

- ✅ `EncodeVarint(io.Writer, uint64) (n int, err error)`  
  Writes a `uint64` as a compact varint to any `io.Writer`. Uses **7 bits per byte**, with the MSB as a continuation flag. Max 10 bytes for `uint64`.

- ✅ `DecodeVarint(io.Reader) (val uint64, err error)`  
  Reads and decodes a varint from any `io.Reader`. Handles multi-byte sequences safely; returns error on malformed input or overflow (>70 bits).

- ✅ `VarintLen(uint64) int`  
  Computes the exact byte length required to encode a value *without* performing I/O — useful for buffer preallocation or size estimation.

## Encoding Scheme

- Each byte holds 7 bits of data; bit 8 (MSB) = `1` if more bytes follow, else `0`.
- Example:  
  - `0x7F` → `1 byte`: `0111_1111`  
  - `0x80` → `2 bytes`: `1000_0000 0000_0001`

## Use Cases

- Network protocols & binary file formats requiring compact integer representation  
- Serialization frameworks (e.g., custom protobuf-like codecs)  
- Embedded systems or bandwidth-constrained environments where space efficiency matters

## Design Notes

- No external dependencies; uses only `io` from the standard library.  
- Thread-safe *per call* (no shared state), but `io.Reader`/`Writer` concurrency must be handled externally.  
- Compatible with standard protobuf varint format (e.g., interoperable with `encoding/binary` or gRPC).
