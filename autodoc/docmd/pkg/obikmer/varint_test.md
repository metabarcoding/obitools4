# Varint Encoding and Decoding Module (`obikmer`)

This Go package implements **variable-length integer encoding/decoding**, commonly used in binary protocols (e.g., Protocol Buffers, SQLite) to efficiently store small integers using fewer bytes.

## Core Features

- **`EncodeVarint(w io.Writer, v uint64) (n int, err error)`**  
  Encodes a `uint64` value into the minimal number of bytes (1–10) using **LEB128-style varint**, writing the result to a writer. Returns bytes written and any I/O error.

- **`DecodeVarint(r io.Reader) (uint64, error)`**  
  Reads and decodes a varint from an `io.Reader`, reconstructing the original `uint64`. Fails on malformed or incomplete data.

- **`VarintLen(v uint64) int`**  
  Computes the exact number of bytes required to encode `v`, without performing I/O.

## Test Coverage

- **Round-trip correctness**: All test values (including edge cases like `0`, powers of two, and max `uint64`) encode → decode back identically.
- **Length validation**: Encoded length matches `VarintLen` predictions exactly (e.g., 127 → 1 byte; 16384 → 3 bytes).
- **Sequence handling**: Multiple varints can be concatenated and decoded in order, preserving data integrity.

## Efficiency & Design

- Uses **7-bit groups per byte**, with the MSB as a continuation flag (`1` = more bytes follow).
- Minimal memory footprint — no allocations beyond buffer I/O.
- Designed for streaming use (e.g., network or file serialization).

## Edge Cases Verified

| Value          | Encoded Length |
|----------------|---------------|
| `0`            | 1 byte        |
| `2⁷−1 = 127`   | 1 byte        |
| `2⁷ = 128`     | 2 bytes       |
| `2¹⁴−1 = 16383`| 2 bytes       |
| `^uint64(0)`   | **10 bytes**  |

