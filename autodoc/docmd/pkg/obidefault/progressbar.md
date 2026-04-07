# Progress Bar Control Module (`obidefault`)

This Go package provides a simple, global mechanism to enable or disable progress bar display across an application.

## Core Functionality

- **`ProgressBar()`**: Returns `true` if progress bars are *enabled* (i.e., when `__no_progress_bar__` is `false`).  
- **`NoProgressBar()`**: Returns the current state of `__no_progress_bar__`, i.e., whether progress bars are *disabled*.  
- **`SetNoProgressBar(b bool)`**: Sets the global flag `__no_progress_bar__`. Passing `true` disables progress bars; passing `false` enables them.  
- **`NoProgressBarPtr()`**: Returns a pointer to the internal `__no_progress_bar__` variable, allowing direct read/write access (e.g., for reflection or UI binding).

## Design Intent

- Centralizes progress bar visibility control in one place.
- Supports both boolean query/set and pointer-based manipulation for flexibility (e.g., CLI flags, config binding).
- Uses a *negative* flag name (`__no_progress_bar__`) internally to default progress bars **on** (i.e., `false` → enabled).

## Usage Example

```go
// Disable progress bars globally:
obidefault.SetNoProgressBar(true)

// Check status:
if !obidefault.ProgressBar() {
    log.Println("Progress bars are disabled.")
}
```

## Notes

- Thread-safety is *not* guaranteed; concurrent access should be externally synchronized.
- The double underscore prefix (`__no_progress_bar__`) signals internal/private usage per Go convention (though not enforced).
