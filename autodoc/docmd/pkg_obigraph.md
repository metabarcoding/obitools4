# `obigraph`: Semantic Overview of Public Features

The `obigraph` package delivers a lightweight, type-safe graph modeling toolkit in Go—optimized for performance and visualization-ready output. Built around two core abstractions (`Graph` and `GraphBuffer`), it supports both static graph construction (for batch processing) and high-throughput streaming ingestion (via buffered channels), while enabling customizable vertex/edge semantics, degree-based filtering, and GML export with visual styling.

---

## Core Graph Type: `Graph[V, T]`

### Generic Structure
- **Type Parameters**:  
  - `V`: Vertex type (arbitrary comparable Go value).  
  - `T`: Edge data payload (e.g., weight, label, metadata).
- **Internal Representation**:  
  - Forward adjacency: `map[V]map[V]T` (outgoing edges).  
  - Reverse adjacency: `map[V]map[V]T` (incoming edges), enabling bidirectional traversal.

### Edge Management
- **Undirected Edges**:  
  - `AddEdge(src, dst V, data T)`: Inserts symmetric links (both directions).  
- **Directed Edges**:  
  - `AddDirectedEdge(src, dst V, data T)`: Inserts one-way link.  
  - `SetAsDirectedEdge(src, dst V)`: Converts existing undirected edge to directed by deleting reverse link.

### Graph Queries
- **Neighbors**:  
  - `Neighbors(v V) []V`: Returns all vertices reachable *from* `v` (successors).  
- **Parents**:  
  - `Parents(v V) []V`: Returns all vertices with edges *to* `v` (predecessors).  
- **Degrees**:  
  - `Degree(v V) int`: Out-degree (size of outgoing adjacency).  
  - `ParentDegree(v V) int`: In-degree (size of incoming adjacency).

### Customization Hooks
- **Vertex Weight**:  
  - `func VertexWeight(v V) float64` (default: constant weight = `1.0`).  
- **Edge Weight**:  
  - `func EdgeWeight(src, dst V) float64` (default: constant weight = `1.0`).  
- **Vertex Labeling**:  
  - `func VertexId(v V) string` (default: `"V%d"` with auto-incrementing index).

### GML Export
- **In-Memory Generation**:  
  - `Gml(w io.Writer, opts ...Option) error`: Renders GML to any writer.  
- **File Output**:  
  - `WriteGmlFile(filename string, opts ...Option) error`: Writes GML to disk.  
- **Styling Options** (via `text/template`):  
  - Directed/undirected mode (`Directed: bool`).  
  - Degree-based filtering (`MinDegree int`): Omits vertices below threshold.  
  - Visual layout:  
    - Shape = `circle` if vertex weight ≥ `Threshold`, else `rectangle`.  
    - Size ∝ sqrt(vertex weight).  

> ⚠️ Errors during template parsing or I/O cause panics (fail-fast design).

---

## Streaming Graph Builder: `GraphBuffer[V, T]`

### Asynchronous Edge Ingestion
- **Channel-Based Protocol**:  
  - Edges are enqueued via `AddEdge(src, dst V)` / `AddDirectedEdge(...)` → pushes to internal buffered channel.  
  - Background goroutine consumes edges and mutates underlying `Graph[V, T]`.  

### Non-Blocking API
- All edge-addition methods return immediately (no synchronization on mutation).  
- Ideal for producer-consumer patterns where multiple goroutines feed edges.

### Lifecycle Management
- **Initialization**:  
  - `NewGraphBuffer(cap int) *GraphBuffer[V, T]`: Starts worker goroutine and allocates channel.  
- **Shutdown**:  
  - `Close()`: Closes ingestion channel → signals worker to terminate gracefully.

### GML Export (Same as `Graph`)
- Supports identical options (`MinDegree`, `Directed`, etc.) via inherited methods.  
- Enables exporting *final* state after streaming completes.

> ⚠️ **Concurrency Note**:  
> - `AddEdge` is *not* safe for concurrent calls without external buffering (e.g., use channel per producer).  
> - The buffer itself handles internal mutation safety via sequential processing.

---

## Use Cases
- **Batch Graph Construction**: `Graph` for offline analysis, static topology generation.  
- **Real-Time Processing**: `GraphBuffer` for event-driven systems (e.g., social feeds, telemetry streams).  
- **Visualization Prep**: GML export supports tools like Graphviz/Cytoscape with minimal styling overhead.
