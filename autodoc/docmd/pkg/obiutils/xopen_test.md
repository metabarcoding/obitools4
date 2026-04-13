# `obiutils` Package: Semantic Overview

The `xopen.go` test suite (via GoCheck) validates utility functions for flexible file/stream I/O in Go. Key features:

- **`IsGzip()`**: Detects gzip compression by inspecting the first two bytes (`0x1f 0x8b`) of a `bufio.Reader`.
- **`Ropen()`**: Unified reader opener supporting:
  - Local files (plain or `.gz`)
  - Standard input (`"-"`) — *note: currently unimplemented in tests*
  - HTTP(S) URLs (via `net/http`)
- **`Wopen()`**: Unified writer opener for:
  - Local files (`".gz"` triggers gzip compression)
  - Standard output via `"-"`
- **`Exists()`**: Checks file/directory existence (supports `~` expansion).
- **`ExpandUser()`**: Expands shell-like paths (`~/...`) to absolute ones.
- **Tested robustness**:
  - Handles missing files, invalid URLs (404), and malformed paths.
  - Validates gzip detection accuracy on both plain and compressed data.

All operations abstract away compression/format details, enabling uniform read/write semantics across local files, pipes (commented out), and remote HTTP resources.
