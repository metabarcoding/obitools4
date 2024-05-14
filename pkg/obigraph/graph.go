package obigraph

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"text/template"

	log "github.com/sirupsen/logrus"
)

type Edge[T any] struct {
	From int
	To   int
	Data *T
}

type Edges[T any] map[int]map[int]*T
type Graph[V any, T any] struct {
	Name         string
	Vertices     *[]V
	Edges        *Edges[T]
	ReverseEdges *Edges[T]
	VertexWeight func(int) float64
	VertexId     func(int) string
	EdgeWeight   func(int, int) float64
}

func NewEdges[T any]() *Edges[T] {
	e := make(map[int]map[int]*T)

	return (*Edges[T])(&e)
}

// AddEdge adds an edge to the graph between two vertices.
//
// Parameters:
// - from: the index of the starting vertex.
// - to: the index of the ending vertex.
// - data: a pointer to the data associated with the edge.
func (e *Edges[T]) AddEdge(from, to int, data *T) {

	fnode, ok := (*e)[from]
	if !ok {
		fnode = make(map[int]*T)
		(*e)[from] = fnode
	}
	fnode[to] = data
}

// NewGraph creates a new graph with the specified name and vertices.
//
// Parameters:
// - name: a string representing the name of the graph.
// - vertices: a slice of vertices of type V.
//
// Returns:
// - Graph[V, T]: the newly created graph.
func NewGraph[V, T any](name string, vertices *[]V) *Graph[V, T] {
	return &Graph[V, T]{
		Name:         name,
		Vertices:     vertices,
		Edges:        NewEdges[T](),
		ReverseEdges: NewEdges[T](),
		VertexWeight: func(i int) float64 {
			return 1.0
		},
		EdgeWeight: func(i, j int) float64 {
			return 1.0
		},
		VertexId: func(i int) string {
			return fmt.Sprintf("V%d", i)
		},
	}
}

// AddEdge adds an edge between two vertices in the graph.
//
// Parameters:
// - from: the index of the starting vertex.
// - to: the index of the ending vertex.
// - data: a pointer to the data associated with the edge.
func (g *Graph[V, T]) AddEdge(from, to int, data *T) {
	lv := len(*g.Vertices)
	if from >= lv || to >= lv {
		log.Errorf("out of bounds vertex index: %d or %d (max: %d)", from, to, lv-1)
	}

	g.Edges.AddEdge(from, to, data)
	g.Edges.AddEdge(to, from, data)
	g.ReverseEdges.AddEdge(to, from, data)
	g.ReverseEdges.AddEdge(from, to, data)
}

// AddDirectedEdge adds a directed edge from one vertex to another in the graph.
//
// Parameters:
// - from: an integer representing the index of the starting vertex.
// - to: an integer representing the index of the ending vertex.
// - data: a pointer to the data associated with the edge.
func (g *Graph[V, T]) AddDirectedEdge(from, to int, data *T) {
	lv := len(*g.Vertices)

	if from >= lv || to >= lv {
		log.Errorf("out of bounds vertex index: %d or %d (max: %d)", from, to, lv-1)
	}

	g.Edges.AddEdge(from, to, data)
	g.ReverseEdges.AddEdge(to, from, data)
}

// SetAsDirectedEdge sets the edge from one vertex to another as directed in the graph.
//
// Parameters:
// - from: an integer representing the index of the starting vertex.
// - to: an integer representing the index of the ending vertex.
func (g *Graph[V, T]) SetAsDirectedEdge(from, to int) {
	lv := len(*g.Vertices)

	if from >= lv || to >= lv {
		log.Errorf("out of bounds vertex index: %d or %d (max: %d)", from, to, lv-1)
	}

	if _, ok := (*g.Edges)[from][to]; ok {
		if _, ok := (*g.Edges)[to][from]; ok {
			delete((*g.Edges)[to], from)
			delete((*g.Edges)[from], to)
		}

		return
	}

	log.Error("no edge from ", from, " to ", to)

}

// Neighbors generates a list of neighbor vertices for a given vertex index in the graph.
//
// Parameters:
// - v: an integer representing the index of the vertex.
// Returns:
// - []int: a list of neighbor vertices.
func (g *Graph[V, T]) Neighbors(v int) []int {
	if neighbors, ok := (*g.Edges)[v]; ok {
		rep := make([]int, 0, len(neighbors))

		for k := range neighbors {
			rep = append(rep, k)
		}

		return rep
	}

	return nil

}

// Degree calculates the degree of a vertex in a graph.
//
// Parameters:
// - v: an integer representing the index of the vertex.
//
// Returns:
// - an integer representing the degree of the vertex.
func (g *Graph[V, T]) Degree(v int) int {
	if neighbors, ok := (*g.Edges)[v]; ok {
		return len(neighbors)
	}
	return 0
}

// Parents returns a list of parent vertices for a given vertex index in the graph.
//
// Parameters:
// - v: an integer representing the index of the vertex.
//
// Returns:
// - []int: a list of parent vertices.
func (g *Graph[V, T]) Parents(v int) []int {
	if parents, ok := (*g.ReverseEdges)[v]; ok {
		rep := make([]int, 0, len(parents))

		for k := range parents {
			rep = append(rep, k)
		}

		return rep
	}

	return nil
}

// ParentDegree calculates the degree of a vertex in a graph by counting the number of its parent vertices.
//
// Parameters:
// - v: an integer representing the index of the vertex.
//
// Returns:
// - an integer representing the degree of the vertex.
func (g *Graph[V, T]) ParentDegree(v int) int {
	if parents, ok := (*g.ReverseEdges)[v]; ok {
		return len(parents)
	}
	return 0

}

type gml_graph[V any, T any] struct {
	Graph       *Graph[V, T]
	As_directed bool
	Min_degree  int
	Threshold   float64
	Scale       int
}

// Gml generates a GML representation of the graph.
//
// as_directed: whether the graph should be treated as directed or undirected.
// threshold: the threshold value.
// scale: the scaling factor.
// string: the GML representation of the graph.
func (g *Graph[V, T]) Gml(as_directed bool, min_degree int, threshold float64, scale int) string {
	//	(*seqs)[1].Count
	var gml bytes.Buffer

	data := gml_graph[V, T]{
		Graph:       g,
		As_directed: as_directed,
		Min_degree:  min_degree,
		Threshold:   threshold,
		Scale:       scale,
	}

	digraphTpl := template.New("gml_digraph")

	digraph := ` {{$context := .}}
				graph [
				comment "{{ if $context.As_directed }}Directed graph{{ else }}Undirected graph{{ end }} {{ Name }}"
				directed {{ if $context.As_directed }}1{{ else }}0{{ end }}

				{{range $index, $data:= $context.Graph.Vertices}}
				{{ if (ge (Degree $index) $context.Min_degree)}}
				node [ 
					id {{$index}}
					graphics [
						type "{{ Shape $index }}"
						h {{ Sqrt (VertexWeight $index) }}
						w {{ Sqrt (VertexWeight $index) }}
					]
				]
				{{ end }}
				{{ end }}

				{{range $source, $data:= $context.Graph.Edges}}
				{{range $target, $edge:= $data}}
				{{ if and (ge $source $context.Min_degree) (ge $target $context.Min_degree) (or $context.As_directed (lt $source $target))}}
				edge [ source {{$source}} 
					   target {{$target}} 
					   color "#00FF00"
					   ]
				{{ end }}
				{{ end }}
				{{ end }}

				]
				`

	tmpl, err := digraphTpl.Funcs(template.FuncMap{
		"Sqrt":         func(i float64) int { return scale * int(math.Floor(math.Sqrt(i))) },
		"Name":         func() string { return g.Name },
		"VertexId":     func(i int) string { return g.VertexId(i) },
		"Degree":       func(i int) int { return g.Degree(i) },
		"VertexWeight": func(i int) float64 { return g.VertexWeight(i) },
		"Shape": func(i int) string {
			if g.VertexWeight(i) >= threshold {
				return "circle"
			} else {
				return "rectangle"
			}
		},
	}).Parse(digraph)

	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(&gml, data)

	if err != nil {
		panic(err)
	}

	return gml.String()

}

// WriteGml writes the GML representation of the graph to an io.Writer.
//
// w: the io.Writer to write the GML representation to.
// as_directed: whether the graph should be treated as directed or undirected.
// threshold: the threshold value.
// scale: the scaling factor.
func (g *Graph[V, T]) WriteGml(w io.Writer, as_directed bool, min_degree int, threshold float64, scale int) {

	_, err := w.Write([]byte(g.Gml(as_directed, min_degree, threshold, scale)))
	if err != nil {
		panic(err)
	}
}

// WriteGmlFile writes the graph in GML format to the specified file.
//
// filename: the name of the file to write the GML representation to.
// as_directed: whether the graph should be treated as directed or undirected.
// threshold: the threshold value.
// scale: the scaling factor.
func (g *Graph[V, T]) WriteGmlFile(filename string, as_directed bool, min_degree int, threshold float64, scale int) {

	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	g.WriteGml(f, as_directed, min_degree, threshold, scale)
}
