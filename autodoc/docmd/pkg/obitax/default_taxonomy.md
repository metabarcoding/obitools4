# ObiTax: Default Taxonomy Management

This Go package (`obitax`) provides utilities for managing a **default taxonomy instance**, enabling centralized configuration and safe fallback behavior.

## Core Features

- ✅ **Singleton-style default taxonomy**: A single global `Taxonomy` instance can be designated as *the* default via `.SetAsDefault()`.

- ✅ **Thread-safe access**: Uses `sync.Mutex` (implicitly via package-level variable usage) to ensure safe concurrent writes when setting the default.

- ✅ **Graceful fallback with `.OrDefault()`**:  
  - If a `Taxonomy` receiver is `nil`, the method automatically substitutes it with the default taxonomy.  
  - Supports optional panic on failure (`panicOnNil`) if no default is defined.

- ✅ **Utility checks**:  
  - `HasDefaultTaxonomyDefined()` → returns whether a default is currently set.  
  - `DefaultTaxonomy()` → retrieves the current global instance (if any).

## Design Intent

- Promotes **configuration reuse** and reduces boilerplate in client code.  
- Supports robustness: avoids nil dereferences by allowing fallback to a globally configured taxonomy.

## Usage Pattern

```go
tax := NewTaxonomy("my-tax")
tax.SetAsDefault() // Now all `nil` receivers will resolve to this instance
result := someNilTax.OrDefault(true) // Uses default; panics only if none exists
```
