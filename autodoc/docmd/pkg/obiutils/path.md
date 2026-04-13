## `obiutils` Package: File Path Utility Functions

This Go package provides two utility functions for manipulating file paths by removing extensions:

### `RemoveAllExt(p string) string`
- **Purpose**: Strips *all* file extensions from a given path (e.g., `/dir/file.tar.gz` → `/dir/file`).  
- **Mechanism**: Iteratively uses `path.Ext()` and `strings.TrimSuffix` to remove extensions from the *full path*, including directory components if they contain dots (though rare).  
- **Use Case**: Useful when you need to sanitize a full path for naming or comparison, regardless of extension stacking.

### `Basename(path string) string`
- **Purpose**: Extracts the base filename *without any extensions* (e.g., `/dir/file.tar.gz` → `file`).  
- **Mechanism**: Uses `filepath.Base()` to get the filename, then iteratively strips extensions via `strings.TrimSuffix`.  
- **Key Difference**: Operates *only on the filename*, not directory parts — safer and more conventional for typical file handling.

Both functions handle multi-extension files (e.g., `.tar.gz`, `.backup.zip`) robustly. They avoid reliance on `strings.LastIndex` or regex, favoring clarity and standard library usage (`path`, `filepath`).  
Designed for portability across Unix-like systems (uses forward slashes), though Windows paths are supported via `filepath.Base`.
