package obikmer

import (
	"bytes"
	"fmt"
	"math"
	"math/bits"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"github.com/daichi-m/go18ds/sets/linkedhashset"
	"github.com/daichi-m/go18ds/stacks/arraystack"
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
	kmersize int
	kmermask uint64
	prevc    uint64
	prevg    uint64
	prevt    uint64
	graph    map[uint64]uint
}

func MakeDeBruijnGraph(kmersize int) *DeBruijnGraph {
	g := DeBruijnGraph{
		kmersize: kmersize,
		kmermask: ^(^uint64(0) << (uint64(kmersize+1) * 2)),
		prevc:    uint64(1) << (uint64(kmersize) * 2),
		prevg:    uint64(2) << (uint64(kmersize) * 2),
		prevt:    uint64(3) << (uint64(kmersize) * 2),
		graph:    make(map[uint64]uint),
	}

	return &g
}

func (g *DeBruijnGraph) KmerSize() int {
	return g.kmersize
}

func (g *DeBruijnGraph) Len() int {
	return len(g.graph)
}

func (g *DeBruijnGraph) MaxLink() int {
	max := uint(0)
	for _, count := range g.graph {
		if count > max {
			max = count
		}
	}

	return int(max)
}

func (g *DeBruijnGraph) LinkSpectrum() []int {
	max := g.MaxLink()
	spectrum := make([]int, max+1)
	for _, count := range g.graph {
		spectrum[int(count)]++
	}

	return spectrum
}

func (g *DeBruijnGraph) FilterMin(min int) {
	umin := uint(min)
	for idx, count := range g.graph {
		if count < umin {
			delete(g.graph, idx)
		}
	}
}

func (g *DeBruijnGraph) Previouses(index uint64) []uint64 {
	rep := make([]uint64, 0, 4)
	index = index >> 2

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

func (g *DeBruijnGraph) MaxNext(index uint64) (uint64, bool) {
	ns := g.Nexts(index)

	if len(ns) == 0 {
		return uint64(0), false
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

	return rep, true
}

func (g *DeBruijnGraph) MaxPath() []uint64 {
	path := make([]uint64, 0, 1000)
	ok := false
	idx := uint64(0)

	idx, ok = g.MaxHead()

	for ok {
		path = append(path, idx)
		idx, ok = g.MaxNext(idx)
	}

	return path
}

func (g *DeBruijnGraph) LongestPath() []uint64 {
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
			idx, ok = g.MaxNext(idx)
		}

		if w > wmax {
			path = lp
			wmax = w
		}
	}

	return path
}

func (g *DeBruijnGraph) LongestConsensus(id string) (*obiseq.BioSequence, error) {
	path := g.LongestPath()
	s := g.DecodePath(path)

	if len(s) > 0 {
		seq := obiseq.MakeBioSequence(
			id,
			[]byte(s),
			"",
		)

		return &seq, nil
	}

	return nil, fmt.Errorf("cannot identify optimum path")
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

func (g *DeBruijnGraph) MaxHead() (uint64, bool) {
	rep := uint64(0)
	max := uint(0)
	found := false
	for k, w := range g.graph {
		if len(g.Previouses(k)) == 0 && w > max {
			rep = k
			found = true
		}
	}

	return rep, found
}

func (g *DeBruijnGraph) DecodeNode(index uint64) string {
	rep := make([]byte, g.kmersize)
	index >>= 2
	for i := g.kmersize - 1; i >= 0; i-- {
		rep[i], _ = decode[index&3]
		index >>= 2
	}

	return string(rep)
}

func (g *DeBruijnGraph) DecodePath(path []uint64) string {
	rep := make([]byte, 0, len(path)+g.kmersize)
	buf := bytes.NewBuffer(rep)

	if len(path) > 0 {
		buf.WriteString(g.DecodeNode(path[0]))

		for _, idx := range path {
			buf.WriteByte(decode[idx&3])
		}
	}

	return buf.String()
}

func (g *DeBruijnGraph) BestConsensus(id string) (*obiseq.BioSequence, error) {
	path := g.MaxPath()
	s := g.DecodePath(path)

	if len(s) > 0 {
		seq := obiseq.MakeBioSequence(
			id,
			[]byte(s),
			"",
		)

		return &seq, nil
	}

	return nil, fmt.Errorf("cannot identify optimum path")
}

func (g *DeBruijnGraph) Weight(index uint64) int {
	val, ok := g.graph[index]
	if !ok {
		val = 0
	}
	return int(val)
}

func (graph *DeBruijnGraph) append(sequence []byte, current uint64) {

	for i := 0; i < len(sequence); i++ {
		current <<= 2
		current &= graph.kmermask
		b := iupac[sequence[i]]
		if len(b) == 1 {
			current |= b[0]
			graph.graph[current] = uint(graph.Weight(current) + 1)
		} else {
			for j := 0; j < len(b); j++ {
				current &= ^uint64(3)
				current |= b[j]

				graph.graph[current] = uint(graph.Weight(current) + 1)
				graph.append(sequence[(i+1):], current)
			}
			return
		}

	}
}

func (graph *DeBruijnGraph) Push(sequence *obiseq.BioSequence) {
	key := uint64(0)
	s := sequence.Sequence()
	init := make([]uint64, 0, 16)
	var f func(start int, key uint64)
	f = func(start int, key uint64) {
		for i := start; i < graph.kmersize; i++ {
			key <<= 2
			b := iupac[s[i]]
			if len(b) == 1 {
				key |= b[0]
			} else {
				for j := 0; j < len(b); j++ {
					key &= ^uint64(3)
					key |= b[j]
					f(i+1, key)
				}
				return
			}
		}
		init = append(init, key&graph.kmermask)
	}

	f(0, key)

	for _, idx := range init {
		graph.append(s[graph.kmersize:], idx)
	}

}

func (graph *DeBruijnGraph) Gml() string {
	buffer := bytes.NewBuffer(make([]byte, 0, 1000))

	buffer.WriteString(
		`graph [
		comment "De Bruijn graph"
		directed 1
		
		`)

	for idx := range graph.graph {
		node := graph.DecodeNode(idx)
		buffer.WriteString(
			fmt.Sprintf("node [ id \"%s\" ]\n", node),
		)
		n := graph.Nexts(uint64(idx))
		if len(n) == 0 {
			idx <<= 2
			idx &= graph.kmermask
			node := graph.DecodeNode(idx)
			buffer.WriteString(
				fmt.Sprintf("node [ id \"%s\" \n label \"%s\" ]\n", node, node),
			)
		}
	}

	for idx, weight := range graph.graph {
		src := graph.DecodeNode(idx)
		label := decode[idx&3]
		idx <<= 2
		idx &= graph.kmermask
		dst := graph.DecodeNode(idx)

		buffer.WriteString(
			fmt.Sprintf(`edge [ source "%s"
				target "%s" 
				color "#00FF00"
				label "%c[%d]"
				graphics		[
					width		%f
					arrow		"last"
					fill		"#007F80"
					]
				]
				
				`, src, dst, label, weight, math.Log(float64(weight))),
		)
	}
	buffer.WriteString("]\n")
	return buffer.String()

}

// fonction tri_topologique(G, V):
//     T <- une liste vide pour stocker l'ordre topologique
//     S <- une pile vide pour stocker les nœuds sans prédécesseurs
//     pour chaque nœud v dans V:
//         si Pred(v) est vide:
//             empiler S avec v
//     tant que S n'est pas vide:
//         nœud <- dépiler S
//         ajouter nœud à T
//         pour chaque successeur s de nœud:
//             supprimer l'arc (nœud, s) de G
//             si Pred(s) est vide:
//                 empiler S avec s
//     si G contient encore des arcs:
//         renvoyer une erreur (le graphe contient au moins un cycle)
//     sinon:
//         renvoyer T (l'ordre topologique)

// A topological sort of the graph.
func (g *DeBruijnGraph) PartialOrder() *linkedhashset.Set[uint64] {
	S := arraystack.New[uint64]()
	T := linkedhashset.New[uint64]()

	for v := range g.graph {
		if len(g.Previouses(v)) == 0 {
			S.Push(v)
		}
	}

	for !S.Empty() {
		v, _ := S.Pop()
		T.Add(v)
		for _, w := range g.Nexts(v) {
			if T.Contains(g.Previouses(w)...) {
				S.Push(w)
			}
		}
	}
	return T
}

// Calculating the hamming distance between two k-mers.
func (g *DeBruijnGraph) HammingDistance(kmer1, kmer2 uint64) int {
	ident := ^((kmer1 & kmer2) | (^kmer1 & ^kmer2))
	ident |= (ident >> 1)
	ident &= 0x5555555555555555 & g.kmermask
	return bits.OnesCount64(ident)
}
