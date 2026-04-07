# `obidefault` Package — Semantic Overview

This minimal Go package provides a centralized, mutable global flag for controlling warning verbosity across an application.

## Core Functionality

- **`__silent_warning__`**:  
  A package-level boolean variable (unexported) that determines whether warnings should be suppressed.

- **`SilentWarning() bool`**:  
  A read-only accessor returning the current state of `__silent_warning__`. Enables safe, non-mutating checks elsewhere in the codebase.

- **`SilentWarningPtr() *bool`**:  
  Returns a pointer to `__silent_warning__`, allowing external code (e.g., CLI parsers, config loaders) to directly mutate the flag — e.g., `*SilentWarningPtr() = true`.

## Design Intent

- **Simplicity & Centralization**:  
  Avoids scattering warning-control logic; provides a single source of truth.

- **Flexibility**:  
  Supports both *read-only* inspection (via `SilentWarning()`) and *global mutation* (via pointer), useful for early initialization phases.

- **Explicit Semantics**:  
  When `SilentWarning()` returns `true`, all warning-generating code *should* suppress output (implementation responsibility lies outside this package).

## Usage Example

```go
// Suppress warnings globally:
*obidefault.SilentWarningPtr() = true

if !obidefault.SilentWarning() {
    log.Println("⚠️ Warning: something happened")
}
```

> **Note**: The double underscore prefix on `__silent_warning__` signals internal/private status, discouraging direct access.
