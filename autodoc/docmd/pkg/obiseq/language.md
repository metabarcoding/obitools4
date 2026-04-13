# Semantic Description of `obiseq` Language Extensions

The `package obiseq` extends the [Gval](https://github.com/PaesslerAG/gval) expression language with domain-specific functions tailored for bioinformatics and data processing. It integrates utility helpers from `obiutils` to provide type-flexible, robust operations over sequences and collections.

## Core Functionalities

- **Data Inspection**:  
  `len`, `ismap`, `isvector` — retrieve size and type information.

- **Aggregation & Comparison**:  
  `min`, `max` — compute extremal values in slices/maps (via `obiutils.Min/Max`).  
  *(Note: commented-out helper functions suggest prior attempts at manual implementations.)*

- **Type Conversion**:  
  `int`, `numeric` (→ float64), `bool`, `string` — safely coerce arbitrary inputs to target types; fail with fatal logs on invalid data.

- **String Manipulation**:  
  `sprintf`, `subspc` (replace spaces with underscores), `replace` (regex-based substitution), and `substr` — support formatting, normalization, and slicing.

- **Sequence Analysis (Bioinformatics)**:  
  `gc`, `gcskew`, and `composition` — compute nucleotide composition metrics for DNA/RNA sequences (`BioSequence`).  
  - `gc`: GC content ratio (excluding ambiguous bases `'o'`)  
  - `gcskew`: `(G−C)/(G+C)` asymmetry measure  
  - `composition`: returns a map of base counts (e.g., `"a":20.0`, `"g":15.0`)

- **Element Access**:  
  `elementof(seq, idx)` — retrieves item at index/key for slices (`[]interface{}`), maps (`map[string]interface{}`), or strings (by byte position).

- **Control Flow**:  
  `ifelse(cond, then_val, else_val)` — conditional branching within expressions.

- **Quality Support**:  
  `qualities(seq)` — extracts per-base quality scores as a float slice from sequencing reads.

## Design Principles

- **Dynamic Typing**: Accepts `...interface{}` arguments for flexibility.
- **Error Handling**: Uses fatal logging (`log.Fatalf`) on conversion failures; returns typed errors for runtime issues.
- **Extensibility**: Built atop `gval.Language`, enabling custom expression evaluation in pipelines (e.g., filtering reads via GC thresholds).

This package serves as a bridge between high-level scripting and low-level biosequence computation, ideal for rule-based filtering or annotation in NGS workflows.
