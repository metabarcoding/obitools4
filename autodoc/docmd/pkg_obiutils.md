# `obiutils` — Semantic Feature Overview

The **`obiutils`** package is a collection of low- and mid-level utilities for numerical computation, string manipulation, file I/O, concurrency control, data conversion, and format detection—specifically designed for bioinformatics pipelines in the OBITools 4 ecosystem. All public APIs are **type-safe**, **well-documented**, and optimized for performance or correctness depending on use case.

## Core Functional Categories

### 🔢 Numerical Utilities
- **`Abs[T constraints.Signed](x T) T`**: Generic absolute value for signed integers and floats (via `golang.org/x/exp/constraints`).  
- **`Min/Max(...)`**: Unified functions accepting scalars, slices, or maps—uses reflection for heterogeneous inputs; returns errors on empty/unsupported types.  
- **`MinMaxSlice[T constraints.Ordered]([]T) (min, max T)`**: Efficient min/max for ordered slices; panics on empty input.  
- **`MinMultiset[T]`**: Lazy-delete min-priority multiset with O(log n) insertion, amortized O(1) minimum access.

### 📦 Data Structures
- **`Set[E comparable]`**: Generic set using `map[E]struct{}` for O(1) membership; supports union, intersection, add/contains/members.  
- **`Vector[T]`, `Matrix[T][][]T`**: Row-major 2D structures with methods:  
  - `.Column(i)`, `.Rows(indices...)`, `.Dim()` (safe for nil/jagged matrices).  
  - `Make2DArray[T]`, `Make2DNumericArray[T](rows, cols int, zeroed bool)` for allocation.

### 🧠 Type Conversion & Validation
- **`InterfaceToString(i interface{}) string`**,  
  `CastableToInt(...)`,  
  `InterfaceToBool(...)` / `Int` / `Float64`: Safe conversions with typed errors (`NotAnInteger`, etc.).  
- **`MapToMapInterface(...)`, `InterfaceToIntMap(...)` / `StringMap`: Converts generic maps to concrete types via reflection.  
- **`InterfaceToStringSlice(...)`**: Normalizes `[]interface{}` or string slices to `[]string`.

### 📄 File & Stream I/O
- **`ReadLines(path string) ([]string, error)`**: Buffered line-by-line file reading.  
- **`Wfile` abstraction** (`OpenWritingFile`, `CompressStream`) with transparent gzip (via `pgzip`), buffering, and append support.  
- **`Ropen/Wopen(...)`**: Unified opener for files/stdin/HTTP/pipes, auto-detecting gzip/xz/zstd/bzip2 via magic bytes.  
- **`DownloadFile(url, path string)`**: Simple HTTP download with progress bar (no retries/timeouts).  
- **`TarFileReader(r io.Reader, path string)`**: Extracts a single file from TAR by exact name match.

### 🔤 String & ASCII Processing
- **`InPlaceToLower([]byte) []byte`**: Zero-copy uppercase→lowercase conversion for ASCII using bitwise OR (`| 32`).  
- **`UnsafeStringFromBytes([]byte) string`, `UnsafeBytes(string) []byte`**: Zero-copy conversions (⚠️ unsafe; no bounds checks).  
- **`AsciiSet[256]bool`**: Predefined sets (`Space`, `Digit`, `Alpha`) + operations (union, intersect) and helpers:  
  - `.FirstWord(...)`, `TrimLeft(s string)` (via method), `RightSplitInTwo(...)`.

### 📏 Memory & Path Utilities
- **`ParseMemSize(s string) (int, error)`**: Parses `"128K"`, `"5MB"` → bytes.  
- **`FormatMemSize(n int) string`**: Formats byte counts as `"1.5K"`, `"2M"` (powers of 1024).  
- **`RemoveAllExt(path string)`, `Basename(path string)`**: Strip *all* extensions from paths (e.g., `"file.tar.gz"` → `"file"`).

### 📡 Format Detection & MIME Handling
- **`HasBOM([]byte) bool, BOMType`**: Detects UTF-8/16/32 byte order marks.  
- **`DropLastLine([]byte) []byte`**: Trims final newline-delimited line (for truncated files).  
- **`RegisterOBIMimeType(...)`**: Extends MIME detection for bioformats (FASTA/FASTQ, CSV, ecoPCR2, GenBank) via regex/magic headers.

### 🔄 Concurrency & Synchronization
- **`AtomicCounter(start int)`**: Thread-safe counter with `Inc()`, `Dec()`, `Value()` (mutex-protected).  
- **`RegisterAPipe/UnregisterPipe()`, `WaitForLastPipe()`**: Lightweight pipeline sync via `sync.WaitGroup` (logs active goroutines).

### 📊 Ranking & Ordering
- **`IntOrder(data []int) []int`, `ReverseIntOrder(...)`: Returns index permutation for ascending/descending sort (original slice unchanged).  
- **`Order[T sort.Interface](data T) []int`: Generic stable index-based sorting.

### 🧪 Testing & Reliability
- All functions include **unit tests** (table-driven, `reflect.DeepEqual`, subtests).  
- Error handling is explicit and typed; logging via Logrus for debugging.  
- No external dependencies beyond `golang.org/x/exp/constraints` (for generics) and optional libraries (`progressbar`, `pgzip`).  
- Designed for portability across Unix/Windows (uses standard library paths).
