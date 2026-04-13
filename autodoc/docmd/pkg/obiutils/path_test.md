# `obiutils` Package: Functional Overview

The `obiutils` package provides utility functions for common file path manipulations in Go. Its current public API includes:

- **`RemoveAllExt(path string) string`**  
  Strips *all* file extensions from a given path, returning the base name without any trailing suffixes (e.g., `.txt`, `.tar.gz`).  
  - Handles paths with no extensions unchanged.  
  - Correctly processes single- and multi-part (e.g., `.tar.gz`) extensions.  
  - Designed for robustness across Unix-like and cross-platform path conventions.

The package currently includes a single unit test suite:

- **`TestRemoveAllExt(t *testing.T)`**  
  Validates the correctness of `RemoveAllExt` using three test cases:  
    • `"path/to/file"` → unchanged (`"path/to/file"`)  
    • `"path/to/file.txt"` → stripped to `"/file"` (→ `"path/to/file"`)  
    • `"path/to/file.tar.gz"` → fully stripped to `"/file"` (→ `"path/to/file"`)  

This ensures reliable behavior for downstream code relying on extension-agnostic path handling—e.g., in build systems, data pipelines, or file-processing tools.
