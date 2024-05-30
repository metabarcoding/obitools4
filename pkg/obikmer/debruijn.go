package obikmer

import (
	"bytes"
	"container/heap"
	"fmt"
	"math"
	"math/bits"
	"os"
	"slices"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	log "github.com/sirupsen/logrus"
)

type KmerIdx32 uint32
type KmerIdx64 uint64
type KmerIdx128 struct {
	Lo uint64
	Hi uint64
}

var iupac = map[byte][]uint64{
	'a': {0},
	'c': {1},
	'g': {2},
	't': {3},
	'u': {3},
	'r': {0, 2},
	'y': {1, 3},
	's': {1, 2},
	'w': {0, 3},
	'k': {2, 3},
	'm': {0, 1},
	'b': {1, 2, 3},
	'd': {0, 2, 3},
	'h': {0, 1, 3},
	'v': {0, 1, 2},
	'n': {0, 1, 2, 3},
}

var decode = map[uint64]byte{
	0: 'a',
	1: 'c',
	2: 'g',
	3: 't',
}

type KmerIdx_t interface {
	KmerIdx32 | KmerIdx64 | KmerIdx128
}

type DeBruijnGraph struct {
	kmersize int    // k-mer size
	kmermask uint64 // mask used to set to 0 the bits that are not in the k-mer
	prevc    uint64 //
	prevg    uint64
	prevt    uint64
	graph    map[uint64]uint // Kmer are encoded as uint64 with 2 bits per character
}

// MakeDeBruijnGraph creates a De Bruijn Graph with the specified k-mer size.
//
// Parameters:
//
//	kmersize int - the size of the k-mers
//
// Returns:
//
//	*DeBruijnGraph - a pointer to the created De Bruijn's Graph
func MakeDeBruijnGraph(kmersize int) *DeBruijnGraph {
	g := DeBruijnGraph{
		kmersize: kmersize,
		kmermask: ^(^uint64(0) << (uint64(kmersize) * 2)), // k-mer mask used to set to 0 the bits that are not in the k-mer
		prevc:    uint64(1) << (uint64(kmersize-1) * 2),
		prevg:    uint64(2) << (uint64(kmersize-1) * 2),
		prevt:    uint64(3) << (uint64(kmersize-1) * 2),
		graph:    make(map[uint64]uint),
	}

	return &g
}

// KmerSize returns the size of the k-mers in the DeBruijn graph.
//
// This function takes no parameters.
// It returns an integer representing the size of the k-mers.
func (g *DeBruijnGraph) KmerSize() int {
	return g.kmersize
}

// Len returns the length of the graph.
//
// This function takes no parameters.
// It returns an integer representing the number of nodes in the graph.
func (g *DeBruijnGraph) Len() int {
	return len(g.graph)
}

// MaxWeight returns the maximum weight of a node from the DeBruijn's Graph.
//
// It iterates over each count in the graph map and updates the max value if the current count is greater.
// Finally, it returns the maximum weight as an integer.
//
// Returns:
// - int: the maximum weight value.
func (g *DeBruijnGraph) MaxWeight() int {
	max := uint(0)
	for _, count := range g.graph {
		if count > max {
			max = count
		}
	}

	return int(max)
}

// WeightSpectrum calculates the weight spectrum of nodes in the DeBruijn's graph.
//
// No parameters.
// Returns an array of integers representing the weight spectrum.
func (g *DeBruijnGraph) WeightSpectrum() []int {
	max := g.MaxWeight()
	spectrum := make([]int, max+1)
	for _, count := range g.graph {
		spectrum[int(count)]++
	}

	return spectrum
}

// FilterMinWeight filters the DeBruijnGraph by removing nodes with weight less than the specified minimum.
//
// min: an integer representing the minimum count threshold.
func (g *DeBruijnGraph) FilterMinWeight(min int) {
	umin := uint(min)
	for idx, count := range g.graph {
		if count < umin {
			delete(g.graph, idx)
		}
	}
}

func (g *DeBruijnGraph) Previouses(index uint64) []uint64 {
	if _, ok := g.graph[index]; !ok {
		log.Panicf("k-mer %s (index %d) is not in graph", g.DecodeNode(index), index)
	}

	rep := make([]uint64, 0, 4)
	index >>= 2

	if _, ok := g.graph[index]; ok {
		rep = append(rep, index)
	}

	key := index | g.prevc
	if _, ok := g.graph[key]; ok {
		rep = append(rep, key)
	}

	key = index | g.prevg
	if _, ok := g.graph[key]; ok {
		rep = append(rep, key)
	}

	key = index | g.prevt
	if _, ok := g.graph[key]; ok {
		rep = append(rep, key)
	}

	return rep
}

func (g *DeBruijnGraph) Nexts(index uint64) []uint64 {
	if _, ok := g.graph[index]; !ok {
		log.Panicf("k-mer %s (index %d) is not in graph", g.DecodeNode(index), index)
	}

	rep := make([]uint64, 0, 4)
	index = (index << 2) & g.kmermask

	if _, ok := g.graph[index]; ok {
		rep = append(rep, index)
	}

	key := index | 1
	if _, ok := g.graph[key]; ok {
		rep = append(rep, key)
	}

	key = index | 2
	if _, ok := g.graph[key]; ok {
		rep = append(rep, key)
	}

	key = index | 3
	if _, ok := g.graph[key]; ok {
		rep = append(rep, key)
	}

	return rep
}

func (g *DeBruijnGraph) MaxNext(index uint64) (uint64, int, bool) {
	ns := g.Nexts(index)

	if len(ns) == 0 {
		return uint64(0), 0, false
	}

	max := uint(0)
	rep := uint64(0)
	for _, idx := range ns {
		w := g.graph[idx]
		if w > max {
			rep = idx
			max = w
		}
	}

	return rep, int(max), true
}

func (g *DeBruijnGraph) Heads() []uint64 {
	rep := make([]uint64, 0, 10)

	for k := range g.graph {
		if len(g.Previouses(k)) == 0 {
			rep = append(rep, k)
		}
	}

	return rep
}

func (g *DeBruijnGraph) MaxHead() (uint64, int, bool) {
	rep := uint64(0)
	max := uint(0)
	found := false
	for k, w := range g.graph {
		if len(g.Previouses(k)) == 0 && w > max {
			rep = k
			max = w
			found = true
		}
	}

	return rep, int(max), found
}

func (g *DeBruijnGraph) MaxPath() []uint64 {
	path := make([]uint64, 0, 1000)
	ok := false
	idx := uint64(0)

	idx, _, ok = g.MaxHead()

	for ok {
		path = append(path, idx)
		idx, _, ok = g.MaxNext(idx)
	}

	return path
}

func (g *DeBruijnGraph) LongestPath(max_length int) []uint64 {
	var path []uint64
	wmax := uint(0)
	ok := true

	starts := g.Heads()
	for _, idx := range starts {
		lp := make([]uint64, 0, 1000)
		ok = true
		w := uint(0)
		for ok {
			nw := g.graph[idx]
			w += nw
			lp = append(lp, idx)
			idx, _, ok = g.MaxNext(idx)
			if max_length > 0 && len(lp) > max_length {
				ok = false
				w = 0
			}
		}

		if w > wmax {
			path = lp
			wmax = w
		}
	}

	return path
}

func (g *DeBruijnGraph) LongestConsensus(id string) (*obiseq.BioSequence, error) {
	//path := g.LongestPath(max_length)
	path := g.HaviestPath()
	s := g.DecodePath(path)

	if len(s) > 0 {
		seq := obiseq.NewBioSequence(
			id,
			[]byte(s),
			"",
		)

		return seq, nil
	}

	return nil, fmt.Errorf("cannot identify optimum path")
}

func (g *DeBruijnGraph) DecodeNode(index uint64) string {
	rep := make([]byte, g.kmersize)
	for i := g.kmersize - 1; i >= 0; i-- {
		rep[i] = decode[index&3]
		index >>= 2
	}

	return string(rep)
}

func (g *DeBruijnGraph) DecodePath(path []uint64) string {
	rep := make([]byte, 0, len(path)+g.kmersize)
	buf := bytes.NewBuffer(rep)

	if len(path) > 0 {
		buf.WriteString(g.DecodeNode(path[0]))

		for _, idx := range path[1:] {
			buf.WriteByte(decode[idx&3])
		}
	}

	return buf.String()
}

func (g *DeBruijnGraph) BestConsensus(id string) (*obiseq.BioSequence, error) {
	path := g.MaxPath()
	s := g.DecodePath(path)

	if len(s) > 0 {
		seq := obiseq.NewBioSequence(
			id,
			[]byte(s),
			"",
		)

		return seq, nil
	}

	return nil, fmt.Errorf("cannot identify optimum path")
}

// Weight returns the weight of the node at the given index in the DeBruijnGraph.
//
// Parameters:
// - index: the index of the node in the graph.
//
// Returns:
// - int: the weight of the node.
func (g *DeBruijnGraph) Weight(index uint64) int {
	val, ok := g.graph[index]
	if !ok {
		val = 0
	}
	return int(val)
}

// append appends a sequence of nucleotides to the DeBruijnGraph.
//
// Parameters:
// - sequence: a byte slice representing the sequence of nucleotides to append.
// - current: the current node in the graph to which the sequence will be appended.
// - weight: the weight of the added nodes.
func (graph *DeBruijnGraph) append(sequence []byte, current uint64, weight int) {

	if len(sequence) == 0 {
		return
	}

	current <<= 2
	current &= graph.kmermask
	b := iupac[sequence[0]]
	current |= b[0]
	graph.graph[current] = uint(graph.Weight(current) + weight)
	graph.append(sequence[1:], current, weight)

	for j := 1; j < len(b); j++ {
		current &= ^uint64(3)
		current |= b[j]
		graph.graph[current] = uint(graph.Weight(current) + weight)
		graph.append(sequence[1:], current, weight)
	}
}

// Push appends a BioSequence to the DeBruijnGraph.
//
// Parameters:
// - sequence: a pointer to a BioSequence containing the sequence to be added.
func (graph *DeBruijnGraph) Push(sequence *obiseq.BioSequence) {
	s := sequence.Sequence() // Get the sequence as a byte slice
	w := sequence.Count()    // Get the weight of the sequence

	var initFirstKmer func(start int, key uint64)

	// Initialize the first k-mer
	// start is the index of the nucleotide in the k-mer to add
	// key is the value of the k-mer index before adding the start nucleotide
	initFirstKmer = func(start int, key uint64) {
		if start == 0 {
			key = 0
		}

		if start < graph.kmersize {
			key <<= 2
			b := iupac[s[start]]

			for _, code := range b {
				key &= ^uint64(3)
				key |= code
				initFirstKmer(start+1, key)
			}
		} else {
			graph.graph[key] = uint(graph.Weight(key) + w)
			graph.append(s[graph.kmersize:], key, w)
		}
	}

	if sequence.Len() > graph.kmersize {
		initFirstKmer(0, 0)
	}
}

func (graph *DeBruijnGraph) Gml() string {
	buffer := bytes.NewBuffer(make([]byte, 0, 1000))

	buffer.WriteString(
		`graph [
		comment "De Bruijn graph"
		directed 1
		
		`)

	nodeidx := make(map[uint64]int)
	nodeid := 0

	for idx := range graph.graph {
		nodeid++
		nodeidx[idx] = nodeid
		n := graph.Nexts(idx)
		p := graph.Previouses(idx)

		if len(n) == 0 || len(p) == 0 {

			node := graph.DecodeNode(idx)
			buffer.WriteString(
				fmt.Sprintf("node [ id \"%d\" \n label \"%s\" ]\n", nodeid, node),
			)
		} else {
			buffer.WriteString(
				fmt.Sprintf("node [ id \"%d\" ]\n", nodeid),
			)
		}
	}

	for idx := range graph.graph {
		srcid := nodeidx[idx]
		weight := graph.Weight(idx)
		n := graph.Nexts(idx)
		for _, dst := range n {
			dstid := nodeidx[dst]
			weight := min(graph.Weight(dst), weight)
			label := decode[dst&3]
			buffer.WriteString(
				fmt.Sprintf(`edge [ source "%d"
					target "%d" 
					color "#00FF00"
					label "%c[%d]"
					graphics		[
						width		%f
						arrow		"last"
						fill		"#007F80"
						]
					]
					
					`, srcid, dstid, label, weight, math.Sqrt(float64(weight))),
			)
		}

	}
	buffer.WriteString("]\n")
	return buffer.String()

}

// WriteGml writes the DeBruijnGraph to a GML file.
//
// filename: the name of the file to write the GML representation to.
// error: an error if any occurs during the file creation or writing process.
func (graph *DeBruijnGraph) WriteGml(filename string) error {

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(graph.Gml())
	return err
}

// Calculating the hamming distance between two k-mers.
func (g *DeBruijnGraph) HammingDistance(kmer1, kmer2 uint64) int {
	ident := ^((kmer1 & kmer2) | (^kmer1 & ^kmer2))
	ident |= (ident >> 1)
	ident &= 0x5555555555555555 & g.kmermask
	return bits.OnesCount64(ident)
}

type UInt64Heap []uint64

func (h UInt64Heap) Len() int           { return len(h) }
func (h UInt64Heap) Less(i, j int) bool { return h[i] < h[j] }
func (h UInt64Heap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *UInt64Heap) Push(x any) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(uint64))
}

func (h *UInt64Heap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (g *DeBruijnGraph) HaviestPath() []uint64 {

	if g.HasCycle() {
		return nil
	}
	// Initialize the distance array and visited set
	distances := make(map[uint64]int)
	visited := make(map[uint64]bool)
	prevNodes := make(map[uint64]uint64)
	heaviestNode := uint64(0)
	heaviestWeight := 0

	queue := &UInt64Heap{}
	heap.Init(queue)

	startNodes := make(map[uint64]struct{})
	for _, n := range g.Heads() {
		startNodes[n] = struct{}{}
		heap.Push(queue, n)
		distances[n] = g.Weight(n)
		prevNodes[n] = 0
		visited[n] = false
	}

	// Priority queue to keep track of nodes to visit
	for len(*queue) > 0 {
		// Get the node with the smallest distance
		currentNode := heap.Pop(queue).(uint64)

		// If the current node has already been visited, skip it
		if visited[currentNode] {
			continue
		}

		// Mark the node as visited
		visited[currentNode] = true
		weight := distances[currentNode]

		// Update the heaviest node
		if weight > heaviestWeight {
			heaviestWeight = weight
			heaviestNode = currentNode
		}

		if currentNode == 0 {
			log.Warn("current node is 0")
		}
		// Update the distance of the neighbors
		nextNodes := g.Nexts(currentNode)
		for _, nextNode := range nextNodes {
			if nextNode == 0 {
				log.Warn("next node is 0")
			}
			weight := g.Weight(nextNode) + distances[currentNode]
			if distances[nextNode] < weight {
				distances[nextNode] = weight
				prevNodes[nextNode] = currentNode
				visited[nextNode] = false
				heap.Push(queue, nextNode)

				// Keep track of the node with the heaviest weight
				if weight > heaviestWeight {
					heaviestWeight = weight
					heaviestNode = nextNode
				}
			}
		}
	}

	log.Infof("Heaviest node: %d [%v]", heaviestNode, heaviestWeight)
	// Reconstruct the path from the start node to the heaviest node found
	heaviestPath := make([]uint64, 0)
	currentNode := heaviestNode
	for _, ok := startNodes[currentNode]; !ok && !slices.Contains(heaviestPath, currentNode); _, ok = startNodes[currentNode] {
		heaviestPath = append(heaviestPath, currentNode)
		//log.Infof("Current node: %d <- %d", currentNode, prevNodes[currentNode])
		currentNode = prevNodes[currentNode]
	}

	if slices.Contains(heaviestPath, currentNode) {
		log.Fatalf("Cycle detected %v -> %v (%v) len(%v)", heaviestPath, currentNode, startNodes, len(heaviestPath))
		return nil
	}

	heaviestPath = append(heaviestPath, currentNode)

	// Reverse the path
	slices.Reverse(heaviestPath)

	return heaviestPath
}

func (g *DeBruijnGraph) HasCycle() bool {
	// Initialize the visited and stack arrays
	visited := make(map[uint64]bool)
	stack := make(map[uint64]bool)

	// Helper function to perform DFS
	var dfs func(node uint64) bool
	dfs = func(node uint64) bool {
		visited[node] = true
		stack[node] = true

		nextNodes := g.Nexts(node)
		for _, nextNode := range nextNodes {
			if !visited[nextNode] {
				if dfs(nextNode) {
					return true
				}
			} else if stack[nextNode] {
				return true
			}
		}
		stack[node] = false
		return false
	}

	// Perform DFS on each node to check for cycles
	for node := range g.graph {
		if !visited[node] {
			if dfs(node) {
				return true
			}
		}
	}
	return false
}
