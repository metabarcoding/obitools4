# `obiutils` — Universal File I/O with Transparent Compression Support

The `xopen`-based package in the `obiutils` module provides a unified interface for reading and writing files, streams, HTTP resources, or command outputs—**transparently handling multiple compression formats**: gzip, xz, zstd, and bzip2.

## Key Functionalities

- **`Ropen(f string)`**  
  Opens a file, stdin (`"-"`), HTTP(S) URL, or shell command (e.g., `"|gzip -dc file.gz"`) for **buffered reading**, auto-detecting compression via magic bytes.

- **`Wopen(f string)` / `WopenFile(...)`**  
  Opens a file or stdout (`"-"`) for **buffered writing**, automatically compressing output based on extension (`.gz`, `.xz`, `.zst`, `.bz2`).

- **Compression Detection**  
  Functions like `IsGzip()`, `IsXz()`, `IsZst()`, and `IsBzip2()` inspect the first bytes of a buffered reader to infer format.

- **Path Utilities**  
  - `ExpandUser(path)` expands POSIX-style paths (`~`, `~/path`) to absolute ones.  
  - `Exists(path)` checks file existence after user expansion.

- **Error Handling**  
  Defines semantic errors: `ErrNoContent`, `ErrDirNotSupported`.

- **Buffered IO**  
  All readers/writers use a default buffer size of `65,536` bytes for performance.

- **Resource Management**  
  `Close()` methods ensure proper cleanup of underlying readers/writers and compression streams.

## Supported Sources & Formats

| Source            | Format(s)              |
|-------------------|------------------------|
| Local files       | plain, `.gz`, `.xz`, `.zst`, `.bz2` |
| Stdin (`"-"`)     | auto-detected          |
| HTTP(S) URLs      | transparent decompression on stream read |
| Pipe commands (`"|cmd"`) | output piped and auto-decompressed |

This abstraction simplifies bioinformatics or data-processing pipelines where input sources vary widely, and compression is common.
