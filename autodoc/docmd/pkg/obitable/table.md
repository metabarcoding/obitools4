# `obitable`: Row-Oriented Data Table for Biological Sequences

The `obitable` package provides a lightweight, row-oriented data table structure (`Table`) for managing biological sequence metadata in Go.

- **Core Types**:
  - `Header`: An ordered column list (alias for `stl4go.Ordered`).
  - `Row`: A flexible map from column names to values (`map[string]interface{}`).
  - `Table`: Holds schema info via `ColType` (column → Go type) and a slice of rows.

- **Row Generators**:
  - `RowFromMap`: Wraps a generic map into a callable row accessor, substituting missing keys with `navalue`.
  - `RowFromBioSeq`: Specialized generator for `obiseq.BioSequence` objects, mapping standard fields (`id`, `sequence`, etc.) and annotations dynamically.

- **Semantic Features**:
  - Supports heterogeneous data types per column (via `reflect.Type`).
  - Enables uniform access to sequence metadata and custom annotations.
  - Designed for interoperability with `obiseq` (OBITools4’s biological sequence module).
  - Facilitates lazy or on-demand row construction—ideal for streaming pipelines.

- **Use Cases**:
  - Converting sequence datasets into tabular formats (e.g., for export, filtering).
  - Building intermediate representations in bioinformatics workflows.
