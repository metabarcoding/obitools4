# Output Compression Control Module

This Go package (`obidefault`) provides a simple, global configuration mechanism for toggling output compression behavior across an application.

## Core Features

- **Global Compression Flag**: A package-level boolean variable `__compress__` (default: `false`) controls whether output should be compressed.
- **Read Access**:  
  - `CompressOutput()` returns the current compression setting as a boolean.
- **Write Access**:  
  - `SetCompressOutput(b bool)` updates the compression flag to a new value.
- **Pointer Access**:  
  - `CompressOutputPtr()` returns a pointer to the internal flag, enabling indirect modification (e.g., for UI bindings or reflection-based updates).

## Design Intent

- Minimal, side-effect-free API.
- Thread-safety *not* guaranteed — intended for use in single-threaded initialization or controlled environments.
- Encapsulation via unexported variable `__compress__`, enforced through accessor functions.

## Typical Usage

```go
// Enable compression globally:
obidefault.SetCompressOutput(true)

if obidefault.CompressOutput() {
    // Apply compression logic (e.g., gzip, brotli)
}
```

## Notes

- The double underscore prefix (`__compress__`) signals internal/private status (convention, not enforced).
- Designed for runtime configurability without recompilation.
