# `obigraph` Package: Semantic Overview

The `obigraph` package provides a generic, type-safe undirected/directed graph implementation in Go. Its core features include:

- **Generic Graph Structure**: Parametrized over vertex type `V` and edge data type `T`, enabling flexible use with arbitrary user-defined types.
- **Bidirectional Edge Tracking**: Maintains both forward (`Edges`) and reverse (`ReverseEdges`) adjacency maps for efficient neighbor/parent queries.
- **Edge Management**:
  - `AddEdge`: Adds an *undirected* edge (inserted in both directions).
  - `AddDirectedEdge`: Adds a *directed* edge (only one direction).
  - `SetAsDirectedEdge`: Converts an existing undirected edge into a directed one by removing the reverse link.
- **Graph Queries**:
  - `Neighbors(v)`: Returns all adjacent vertices (outgoing in directed case).
  - `Parents(v)`: Returns incoming neighbors via reverse adjacency.
  - `Degree(v)` / `ParentDegree(v)`: Compute vertex degrees (total or incoming).
- **Customizable Vertex/Edge Properties**:
  - `VertexWeight`, `EdgeWeight`: Funcs to assign weights (default: constant weight = 1.0).
  - `VertexId`: Custom vertex label generator (default: `"V%d"`).

- **GML Export**:
  - `Gml(...)` / `WriteGml(...)`: Generates or writes a Graph Modelling Language (GML) representation.
  - Supports directed/undirected modes, degree-based filtering (`min_degree`), and visual styling:
    - Vertex shape: `circle` if weight ≥ threshold, else `rectangle`.
    - Size scaled by square root of vertex weight.
  - Uses Go’s `text/template` for rendering.

- **File I/O**: Directly writes GML to file via `WriteGmlFile(...)`.

- **Logging & Safety**: Uses Logrus for bounds-checking errors; panics on template parsing/writing failures.

The package is designed for lightweight, high-performance graph modeling and visualization-ready export.
