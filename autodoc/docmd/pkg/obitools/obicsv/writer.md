# CSV Export Functionality in `obicsv` Package

The `obicsv` package provides utilities for efficiently writing structured data (e.g., sequence annotations) to CSV format, supporting parallel processing and streaming.

- **`FormatCVSBatch()`**: Converts a batch of CSV records (`CSVRecordBatch`) into an in-memory buffer, using the provided header and a placeholder for missing values (`navalue`). It prepends the header only once (for batch order 0).

- **`WriteCSV()`**: Writes a CSV-formatted stream from an `ICSVRecord` iterator to any `io.WriteCloser`. It supports:
  - Compression (via `obiutils.CompressStream`)
  - Parallel workers for batch processing (`ParallelWorkers()`)
  - Chunked writing via `obiformats.WriteFileChunk`
  
- **`WriteCSVToStdout()` / `WriteCSVToFile()`**: Convenience wrappers:
  - Outputs to stdout (`os.Stdout`)
  - Writes to a file (with `O_WRONLY`, optional append/truncate)

- **Key design features**:
  - Non-blocking, concurrent processing using goroutines
  - Graceful shutdown via `WaitAndClose()` and channel signaling
  - Robust handling of missing/invalid values (falls back to `navalue`)
  
- **Dependencies**: Leverages internal packages for iteration (`obiitercsv`), data formats (`obiformats`), and utilities (`obiutils`, `logrus` logging).
