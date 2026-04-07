# BioSequence Attribute Management API

This Go package (`obiseq`) provides a rich set of methods for managing metadata and structural attributes associated with biological sequences (`BioSequence`). Below is a semantic overview of the core functionalities:

- **Key Discovery & Existence Checks**:  
  - `Keys()` and `AttributeKeys()` return all attribute names (optionally excluding container/statistics fields or the `"definition"` key).  
  - `HasAttribute(key)` verifies presence of a given attribute (including standard fields: `"id"`, `"sequence"`, `"qualities"`).

- **Generic Attribute Access**:  
  - `GetAttribute(key)` retrieves any attribute value (as `interface{}`), with thread-safe locking.  
  - `SetAttribute(key, value)` assigns values to attributes (including automatic conversion for `"id"`, `"sequence"` and `"qualities"`).

- **Typed Attribute Retrieval**:  
  - Type-specific getters (`GetIntAttribute`, `GetFloatAttribute`, `GetStringAttribute`, etc.) ensure safe conversion and *auto-upgrade* of stored values (e.g., string `"42"` → integer `42`).  
  - Supports maps (`GetIntMap`, `GetStringMap`) and slices (`GetIntSlice`).

- **Convenience & Domain-Specific Helpers**:  
  - `Count()` / `SetCount()`: manage observation frequency (default = 1).  
  - OBITag indexing: `OBITagRefIndex()` / `SetOBITagRefIndex()`, and geometry variants (`geomref`). Supports flexible input map types with dynamic conversion.  
  - Coordinate & landmark support: `GetCoordinate()` / `SetCoordinate()`, and `landmark_id`-based operations (`IsALandmark()`, `GetLandmarkID()`).

All methods are designed for robustness: they handle type conversions gracefully, use locking to ensure concurrency safety, and provide fallbacks (e.g., default count = 1). The API abstracts internal storage (`annotations` map) while exposing a clean, consistent interface for sequence annotation manipulation.
