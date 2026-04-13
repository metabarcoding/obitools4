# `obiutils` Package: File and Stream Writing Utilities

The `obiutils` package provides a unified abstraction for writing data to files or streams, with optional gzip compression and buffered I/O.

## Core Type: `Wfile`

- Encapsulates a write-ready output stream (`io.WriteCloser`).
- Supports both **compressed** (gzip) and uncompressed modes.
- Uses `bufio.Writer` for efficient buffered writes.

## Key Functions

### `OpenWritingFile(name string, compressed bool, append bool) (*Wfile, error)`
- Opens a file for writing.
  - `compressed`: enables gzip compression via `pgzip`.
  - `append`: if true, writes at end of file (`os.O_APPEND`).
- Returns a ready-to-use `*Wfile`.

### `CompressStream(out io.WriteCloser, compressed bool, close bool) (*Wfile, error)`
- Wraps an arbitrary `io.WriteCloser` (e.g., HTTP response, pipe) in buffered/compressed I/O.
  - `close`: if true, the underlying writer is closed on `.Close()`.

## Methods

- **`Write(p []byte)` / `WriteString(s string)`**:  
  Buffered writes to the underlying stream (transparently compressed if enabled).

- **`Close()`**:  
  - Flushes the buffer.
  - Closes gzip writer (if compressed).
  - Closes underlying file/stream *only if* `close == true`.

## Design Highlights

- **Transparent compression**: Uses high-performance `pgzip` for parallel gzip.
- **Resource control**: Explicit flag (`close`) prevents premature closure of shared writers (e.g., in pipelines).
- **Efficiency**: Double buffering via `bufio.Writer` + gzip stream.
