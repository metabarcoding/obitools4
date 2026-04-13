# `obitable`: Row-Oriented Data Table for Biological Sequences

The `obitable` package provides a lightweight, row-oriented data table structure (`Table`) for managing biological sequence metadata in Go. It is designed to support heterogeneous, schema-flexible tabular representations of sequences while maintaining strong interoperability with `obiseq`, the core biological sequence module in OBITools4.

## Core Types

- **`Header`**: An ordered list of column names (alias for `stl4go.Ordered`). It defines the schema’s *column order* but not types.
- **`Row`**: A flexible, map-like structure (`map[string]interface{}`) representing a single record. Values may be of any Go type.
- **`Table`**: Encapsulates both a `Header`, column-type metadata (`ColType map[string]reflect.Type`), and an ordered slice of `Row`s. Enforces type consistency *per column* across rows.

## Row Generators (Lazy/On-Demand Construction)

- **`RowFromMap(map[string]interface{}, interface{}) RowFunc`**  
  Returns a callable `func(string) interface{}` (i.e., a *row accessor*). For any column name, it retrieves the corresponding value from the input map; missing keys are replaced by a configurable default (`navalue`, typically `nil`). Enables efficient wrapping of generic maps as row-like functions.

- **`RowFromBioSeq(seq obiseq.BioSequence, navalue interface{}) RowFunc`**  
  Constructs a row accessor specialized for `obiseq.BioSequence`. Maps standard fields (`id`, `description`, `sequence`, etc.) and dynamically extracts all sequence annotations (e.g., `qualifiers` in FASTA/FASTQ) as column entries. Missing fields default to `navalue`.

## Semantic Capabilities

- **Heterogeneous column types**: Each column may hold values of any Go type (e.g., `string`, `int64`, `[]byte`), with runtime type tracking via `ColType`.
- **Uniform metadata access**: Enables seamless integration of sequence identifiers, raw sequences, and rich annotation sets (e.g., taxonomy IDs, quality scores).
- **Streaming-friendly**: Row generators avoid materializing full row maps until needed—ideal for large-scale pipeline processing.
- **Interoperability**: Built explicitly to work with `obiseq` and future extensions of OBITools4.

## Public API Summary

| Function / Type | Purpose |
|-----------------|---------|
| `NewTable(header Header, colType map[string]reflect.Type) *Table` | Instantiate a new table with schema. |
| `Append(t *Table, row Row)` | Append one fully materialized row to the table. |
| `AppendFunc(t *Table, f RowFunc)` | Append a row via lazy accessor (no intermediate map). |
| `RowFromMap(...), RowFromBioSeq(...)` | Create reusable row accessors for map-based or sequence-backed data. |
| `ToMap(row Row) map[string]interface{}` | Materialize a row as a plain Go map. |
| `ColType(t *Table) map[string]reflect.Type` | Expose column type metadata. |
| `Header(t *Table) Header` | Retrieve the table’s ordered header (column names). |
| `Rows(t *Table) []Row` | Access all rows as a slice (for iteration/export). |

> **Note**: All public functions operate on `Table`, `Header`, and `Row` types. Internal helpers (e.g., type-checking utilities) are not exposed.
