package obigraph

import (
	"io"
	"os"
)

type GraphBuffer[V, T any] struct {
	Graph   *Graph[V, T]
	Channel chan Edge[T]
}

// NewGraphBuffer creates a new GraphBuffer with the given name and vertices.
//
// Parameters:
// - name: the name of the GraphBuffer.
// - vertices: a slice of vertices to initialize the GraphBuffer.
//
// Returns:
// - GraphBuffer[V, T]: the newly created GraphBuffer.
func NewGraphBuffer[V, T any](name string, vertices *[]V) *GraphBuffer[V, T] {
	buffer := GraphBuffer[V, T]{
		Graph:   NewGraph[V, T](name, vertices),
		Channel: make(chan Edge[T]),
	}

	go func() {
		for edge := range buffer.Channel {
			buffer.Graph.AddEdge(edge.From, edge.To, edge.Data)
		}
	}()

	return &buffer
}

// AddEdge adds an edge to the GraphBuffer.
//
// Parameters:
// - from: the index of the starting vertex.
// - to: the index of the ending vertex.
// - data: a pointer to the data associated with the edge.
func (g *GraphBuffer[V, T]) AddEdge(from, to int, data *T) {
	g.Channel <- Edge[T]{
		From: from,
		To:   to,
		Data: data,
	}
}

// AddDirectedEdge adds a directed edge from one vertex to another in the GraphBuffer.
//
// Parameters:
// - from: the index of the starting vertex.
// - to: the index of the ending vertex.
// - data: a pointer to the data associated with the edge.
func (g *GraphBuffer[V, T]) AddDirectedEdge(from, to int, data *T) {
	g.Channel <- Edge[T]{
		From: from,
		To:   to,
		Data: data,
	}
}

// Gml generates a GML representation of the graph.
//
// as_directed: whether the graph should be treated as directed or undirected.
// min_degree: the minimum degree of vertices to include in the GML representation.
// threshold: the threshold value.
// scale: the scaling factor.
// string: the GML representation of the graph.
func (g *GraphBuffer[V, T]) Gml(as_directed bool, min_degree int, threshold float64, scale int) string {
	return g.Graph.Gml(as_directed, min_degree, threshold, scale)
}

func (g *GraphBuffer[V, T]) WriteGmlFile(filename string, as_directed bool, min_degree int, threshold float64, scale int) {

	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	g.WriteGml(f, as_directed, min_degree, threshold, scale)
}

// WriteGml writes the GML representation of the graph to an io.Writer.
//
// w: the io.Writer to write the GML representation to.
// as_directed: whether the graph should be treated as directed or undirected.
// min_degree: the minimum degree of vertices to include in the GML representation.
// threshold: the threshold value.
// scale: the scaling factor.
func (g *GraphBuffer[V, T]) WriteGml(w io.Writer, as_directed bool, min_degree int, threshold float64, scale int) {
	_, err := w.Write([]byte(g.Gml(as_directed, min_degree, threshold, scale)))
	if err != nil {
		panic(err)
	}
}

// Close closes the GraphBuffer by closing its channel.
//
// No parameters.
func (g *GraphBuffer[V, T]) Close() {
	close(g.Channel)
}
