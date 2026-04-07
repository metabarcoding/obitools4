# `obitax` Package: Taxonomic Identifier Handling

The `obitax` package provides a lightweight, type-safe abstraction for handling taxonomic identifiers (`Taxid`) in the OBITools4 ecosystem.

- **`Taxid` type**: A pointer to a string, representing an opaque taxonomic ID (e.g., NCBI TaxID).
- **`TaxidFactory`**: A factory for constructing `Taxid`s from strings or integers, enforcing validation and normalization.

Key features:
- **Code prefix enforcement**: `FromString` validates that the input string starts with a required taxonomy code (e.g., `"tx"`), returning an error otherwise.
- **String parsing**: Automatically strips leading whitespace and extracts the suffix after `':'`.
- **Alphabet filtering**: Uses an ASCII set to extract only valid characters (e.g., digits), ensuring clean, standardized IDs.
- **String interning**: Internally uses `Innerize` (via `InnerString`) to deduplicate strings—improving memory efficiency and comparison speed.
- **Type safety**: `Taxid` is a distinct type (not raw string), reducing misuse and enabling future extension.

Supported conversions:
- `FromString(string)`: Parses `"tx:12345"` → internalized `"12345"`.
- `FromInt(int)`: Converts e.g., `12345` → internalized `"12345"`.

Designed for high-performance pipelines where many taxonomic IDs are processed and reused.
