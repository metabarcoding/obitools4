# `obigraph.GraphBuffer` Feature Overview

The `GraphBuffer[V, T]` type provides a **thread-safe graph construction interface** using buffered edge insertion via Go channels.

- **Asynchronous Edge Addition**: Edges are enqueued through a `chan Edge[T]`, processed in the background by a goroutine that updates an underlying static graph (`Graph[V, T]`).  
- **Non-blocking API**: `AddEdge` and `AddDirectedEdge` are non-synchronous — they send to the channel without waiting for graph mutation, enabling high-throughput edge ingestion.  
- **Graph Initialization**: `NewGraphBuffer` initializes both the graph and a dedicated worker goroutine to consume edges.  
- **GML Export Support**: Full support for exporting the final graph in [Graph Modelling Language (GML)](https://en.wikipedia.org/wiki/Graph_Modelling_Language), with optional filtering (`min_degree`) and layout parameters (`threshold`, `scale`).  
- **File & Stream Output**: Methods `WriteGml` and `WriteGmlFile` allow writing GML to any `io.Writer`, including files.  
- **Resource Cleanup**: The explicit `Close()` method terminates the worker goroutine by closing the channel, ensuring clean shutdown.  
- **Generic Design**: Fully generic over vertex (`V`) and edge data types (`T`), supporting arbitrary value semantics.  

> ⚠️ **Note**: The buffer is *not* safe for concurrent `AddEdge` calls without external synchronization beyond channel semantics.  
> ✅ Ideal for producer-consumer patterns where edges are streamed from multiple goroutines into a single graph.
