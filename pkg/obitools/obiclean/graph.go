package obiclean

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path"
	"sort"
	"sync"
	"text/template"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obialign"
	"github.com/schollz/progressbar/v3"
)

type Ratio struct {
	Sample string
	From   int
	To     int
	CFrom  int
	CTo    int
	Pos    int
	Length int
}

type Edge struct {
	Father  int
	From    byte
	To      byte
	Pos     int
	NucPair int
	Dist    int
}

func makeEdge(father, dist, pos int, from, to byte) Edge {
	return Edge{
		Father:  father,
		Dist:    dist,
		Pos:     pos,
		From:    from,
		To:      to,
		NucPair: nucPair(from, to),
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func minMax(x, y int) (int, int) {
	if x < y {
		return x, y
	}
	return y, x

}

// It takes a filename and a 2D slice of floats pruduced during graph building,
// and writes a CSV file with the first column being the
// first nucleotide, the second column being the second nucleotide, and the third column being the
// ratio
func EmpiricalDistCsv(filename string, data [][]Ratio) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	pbopt := make([]progressbar.Option, 0, 5)
	pbopt = append(pbopt,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(15),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionSetDescription("[Save CSV stat ratio file"),
	)

	bar := progressbar.NewOptions(len(data), pbopt...)

	fmt.Fprintln(file, "Sample,From,To,Weight_from,Weight_to,Count_from,Count_to,Position,length")
	for code, dist := range data {
		a1, a2 := intToNucPair(code)
		for _, ratio := range dist {
			fmt.Fprintf(file, "%s,%c,%c,%d,%d,%d,%d,%d,%d\n",
				ratio.Sample,
				a1, a2,
				ratio.From,
				ratio.To,
				ratio.CFrom,
				ratio.CTo,
				ratio.Pos,
				ratio.Length)
		}
		bar.Add(1)
	}
}

// It takes a slice of sequences, a sample name and a statistical threshold and returns a string
// containing a GML representation of the graph
func Gml(seqs *[]*seqPCR, sample string, statThreshold int) string {
	//	(*seqs)[1].Count
	var dot bytes.Buffer
	digraphTpl := template.New("gml_digraph")
	digraph := `graph [
		comment "Obiclean graph for sample {{ Name }}"
		directed 1
		{{range $index, $data:= .}}
		{{ if or $data.Edges (gt $data.SonCount 0)}}
	node [ id {{$index}} 
			graphics [
				type "{{ Shape $data.Count }}"
				fill "{{ if and (gt $data.SonCount 0) (not $data.Edges)}}#0000FF{{  else }}#00FF00{{ end }}"
				h {{ Sqrt $data.Count }}
				w {{ Sqrt $data.Count }}
			]
		     
         ]
		 {{ end }}
		 {{ end }}

		 {{range $index, $data:= .}}
				{{range $i, $edge:= $data.Edges}}
	edge [ source {{$index}} 
	       target {{$edge.Father}} 
		   color "{{ if gt (index $data.Edges $i).Dist 1 }}#FF0000{{  else }}#00FF00{{ end }}"
		   label "{{(index $data.Edges $i).Dist}}"
		   ]
				{{ end }}
				{{ end }}
	] 
		 
			`

	tmpl, err := digraphTpl.Funcs(template.FuncMap{
		"Sqrt": func(i int) int { return 3 * int(math.Floor(math.Sqrt(float64(i)))) },
		"Name": func() string { return sample },
		"Shape": func(i int) string {
			if i >= statThreshold {
				return "circle"
			} else {
				return "rectangle"
			}
		},
	}).Parse(digraph)

	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(&dot, *seqs)

	if err != nil {
		panic(err)
	}

	return dot.String()
}

func SaveGMLGraphs(dirname string,
	samples map[string]*[]*seqPCR,
	statThreshold int,
) {

	if stat, err := os.Stat(dirname); err != nil || !stat.IsDir() {
		// path does not exist or is not directory
		os.RemoveAll(dirname)
		err := os.Mkdir(dirname, 0755)

		if err != nil {
			log.Panicf("Cannot create directory %s for saving graphs", dirname)
		}
	}

	pbopt := make([]progressbar.Option, 0, 5)
	pbopt = append(pbopt,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(15),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionSetDescription("[Save GML Graph files]"),
	)

	bar := progressbar.NewOptions(len(samples), pbopt...)

	for name, seqs := range samples {

		file, err := os.Create(path.Join(dirname,
			fmt.Sprintf("%s.gml", name)))

		if err != nil {
			fmt.Println(err)
		}

		file.WriteString(Gml(seqs, name, statThreshold))
		file.Close()

		bar.Add(1)
	}

}

func nucPair(a, b byte) int {

	n1 := 0
	switch a {
	case 'a':
		n1 = 1
	case 'c':
		n1 = 2
	case 'g':
		n1 = 3
	case 't':
		n1 = 4
	}

	n2 := 0
	switch b {
	case 'a':
		n2 = 1
	case 'c':
		n2 = 2
	case 'g':
		n2 = 3
	case 't':
		n2 = 4
	}

	return n1*5 + n2

}

func intToNucPair(code int) (a, b byte) {
	var decode = []byte{'-', 'a', 'c', 'g', 't'}
	c1 := code / 5
	c2 := code - c1*5

	return decode[c1], decode[c2]
}

func reweightSequences(seqs *[]*seqPCR) {

	for _, node := range *seqs {
		node.Weight = node.Count
	}

	//var rfunc func(*seqPCR)

	rfunc := func(node *seqPCR) {
		node.AddedSons = 0
		nedges := len(node.Edges)
		if nedges > 0 {
			swf := 0.0

			for k := 0; k < nedges; k++ {
				swf += float64((*seqs)[node.Edges[k].Father].Count)
			}

			for k := 0; k < nedges; k++ {
				father := (*seqs)[node.Edges[k].Father]
				father.Weight += int(math.Round(float64(node.Weight) * float64(father.Count) / swf))
				father.AddedSons++
				// log.Println(father.AddedSons, father.SonCount)
			}

		}
	}

	for _, node := range *seqs {
		if node.SonCount == 0 {
			rfunc(node)
		}
	}

	for done := true; done; {
		done = false
		for _, node := range *seqs {
			if node.SonCount > 0 && node.SonCount == node.AddedSons {
				// log.Println(node.AddedSons, node.SonCount)
				rfunc(node)
				done = true
			}
		}
	}
}

func buildSamplePairs(seqs *[]*seqPCR, workers int) int {
	nseq := len(*seqs)
	running := sync.WaitGroup{}

	linePairs := func(i int) {

		son := (*seqs)[i]

		for j := i + 1; j < nseq; j++ {
			father := (*seqs)[j]
			if father.Count > son.Count {
				d, pos, a1, a2 := obialign.D1Or0(son.Sequence, father.Sequence)
				if d > 0 {
					son.Edges = append(son.Edges, makeEdge(j, d, pos, a2, a1))
					father.SonCount++

				}
			}
		}

	}

	lineChan := make(chan int)

	ff := func() {
		for i := range lineChan {
			linePairs(i)
		}
		running.Done()
	}

	running.Add(workers)

	for i := 0; i < workers; i++ {
		go ff()
	}

	go func() {
		for i := 0; i < nseq; i++ {
			lineChan <- i
		}
		close(lineChan)
	}()

	np := nseq * (nseq - 1) / 2

	running.Wait()

	reweightSequences(seqs)

	return np
}

func extendSimilarityGraph(seqs *[]*seqPCR, step int, workers int) int {
	nseq := len(*seqs)
	running := sync.WaitGroup{}

	linePairs := func(matrix *[]uint64, i int) {
		son := (*seqs)[i]
		for j := i + 1; j < nseq; j++ {
			father := (*seqs)[j]
			d, _, _, _ := obialign.D1Or0(son.Sequence, father.Sequence)

			if d < 0 {
				lcs, lali := obialign.FastLCSScore(son.Sequence, father.Sequence,
					step,
					matrix)
				d := (lali - lcs)
				if lcs >= 0 && d <= step && step > 0 {
					son.Edges = append(son.Edges, makeEdge(j, d, -1, '-', '-'))
					father.SonCount++
					//a, b := minMax((*seqs)[i].Count, (*seqs)[j].Count)
				}
			}

		}
	}

	lineChan := make(chan int)
	// idxChan := make(chan [][]Ratio)

	ff := func() {
		var matrix []uint64
		
		for i := range lineChan {
			linePairs(&matrix, i)
		}

		running.Done()
	}

	running.Add(workers)

	for i := 0; i < workers; i++ {
		go ff()
	}

	go func() {
		for i := 0; i < nseq; i++ {
			if len((*seqs)[i].Edges) == 0 {
				lineChan <- i
			}
		}
		close(lineChan)
	}()

	running.Wait()
	np := nseq * (nseq - 1) / 2
	return np
}

func FilterGraphOnRatio(seqs *[]*seqPCR, ratio float64) {
	for _, s1 := range *seqs {
		c1 := float64(s1.Weight)
		e := s1.Edges
		j := 0
		for i, s2 := range e {
			e[j] = e[i]
			if (c1 / float64((*seqs)[s2.Father].Weight)) <= math.Pow(ratio, float64(e[i].Dist)) {
				j++
			} else {
				(*seqs)[s2.Father].SonCount--
			}
		}
		s1.Edges = e[0:j]
	}
}

// sortSamples sorts the sequences in each sample by their increasing count
func sortSamples(samples map[string]*([]*seqPCR)) {

	for _, s := range samples {
		sort.SliceStable(*s, func(i, j int) bool {
			return (*s)[i].Count < (*s)[j].Count
		})
	}

}

func ObicleanStatus(seq *seqPCR) string {
	if len(seq.Edges) == 0 {
		if seq.SonCount > 0 {
			return "h"
		} else {
			return "s"
		}
	} else {
		return "i"
	}
}

func EstimateRatio(samples map[string]*[]*seqPCR, minStatRatio int) [][]Ratio {
	ratio := make([][]Ratio, 25)

	for name, seqs := range samples {

		for _, seq := range *seqs {
			for _, edge := range seq.Edges {
				father := (*seqs)[edge.Father]
				if father.Weight >= minStatRatio && edge.Dist == 1 {
					ratio[edge.NucPair] = append(ratio[edge.NucPair], Ratio{name, father.Weight, seq.Weight, father.Count, seq.Count, edge.Pos, father.Sequence.Length()})
				}
			}

		}
	}

	return ratio

}

func BuildSeqGraph(samples map[string]*[]*seqPCR,
	maxError, workers int) {

	sortSamples(samples)

	npairs := 0
	for _, seqs := range samples {
		nseq := len(*seqs)
		npairs += nseq * (nseq - 1) / 2
	}

	pbopt := make([]progressbar.Option, 0, 5)
	pbopt = append(pbopt,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(15),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionSetDescription("[One error graph]"),
	)

	bar := progressbar.NewOptions(npairs, pbopt...)
	for _, seqs := range samples {
		np := buildSamplePairs(seqs, workers)

		bar.Add(np)
	}

	if maxError > 1 {
		pbopt = make([]progressbar.Option, 0, 5)
		pbopt = append(pbopt,
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionSetWidth(15),
			progressbar.OptionShowIts(),
			progressbar.OptionSetPredictTime(true),
			progressbar.OptionSetDescription("[Adds multiple errors]"),
		)

		bar = progressbar.NewOptions(npairs, pbopt...)

		for _, seqs := range samples {
			np := extendSimilarityGraph(seqs, maxError, workers)
			bar.Add(np)
		}
	}
}
