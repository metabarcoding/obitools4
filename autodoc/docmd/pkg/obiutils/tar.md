# `TarFileReader` — Semantic Description

The function `TarFileReader`, defined in the Go package `obiutils`, provides a targeted extraction capability for files within a TAR archive.

- **Input**:  
  - `file`: A generic reader (`*Reader`) implementing the standard Go `io.Reader` interface — typically wrapping an archive file or stream.  
  - `path`: A string specifying the *exact* path (relative to archive root) of the desired file inside the TAR.

- **Core Logic**:  
  - Instantiates a `tar.Reader` from the provided input stream.  
  - Iterates sequentially over TAR entries using `Next()`.  
  - Compares each entry’s header name (`header.Name`) with the requested `path`.

- **Output**:  
  - On match: Returns a pointer to the *current* `tar.Reader`, positioned at the start of the requested file’s content (ready for subsequent reads).  
  - On failure: Returns `nil` and a formatted error `"file not found: <path>"`.

- **Semantics**:  
  - Enables *lazy*, on-demand access to a specific file inside a TAR archive — without decompressing the entire structure.  
  - Assumes exact path matching (no globbing, wildcards, or directory traversal).  
  - Does *not* handle symbolic links, hardlinks, or nested archives — only plain file entries.

- **Use Case**:  
  Ideal for lightweight tools that need to inspect or extract a single known file from large TAR archives (e.g., config files, manifests), minimizing memory and I/O overhead.
