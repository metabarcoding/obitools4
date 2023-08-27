package obirefidx

import (
	"fmt"
	"log"
	"sort"
	"sync"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitax"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiutils"
)

func GeomIndexSesquence(seqidx int,
	references obiseq.BioSequenceSlice,
	taxa *obitax.TaxonSet,
	taxo *obitax.Taxonomy) map[int]string {

	sequence := references[seqidx]
	location := sequence.GetCoordinate()

	if location == nil {
		log.Fatalf("Sequence %s does not have a coordinate", sequence.Id())
	}

	seq_dist := make([]float64, len(references))

	var wg sync.WaitGroup

	for i, ref := range references {
		wg.Add(1)
		go func(i int, ref *obiseq.BioSequence) {
			defer wg.Done()
			reflocation := ref.GetCoordinate()
			if reflocation == nil {
				log.Fatalf("Sequence %s does not have a coordinate", ref.Id())
			}
			d := 0.0
			for i, x := range location {
				diff := float64(x - reflocation[i])
				d += diff * diff
			}
			seq_dist[i] = d
		}(i, ref)
	}

	wg.Wait()

	order := obiutils.Order(sort.Float64Slice(seq_dist))

	lca := (*taxa)[seqidx]

	index := make(map[int]string)
	index[0.0] = fmt.Sprintf(
		"%d@%s@%s",
		lca.Taxid(),
		lca.ScientificName(),
		lca.Rank())

	old_dist := 0.0
	for _, o := range order {
		new_lca, _ := lca.LCA((*taxa)[o])
		if new_lca.Taxid() != lca.Taxid() || seq_dist[o] != old_dist {
			lca = new_lca
			old_dist = seq_dist[o]
			index[int(seq_dist[o])] = fmt.Sprintf(
				"%d@%s@%s",
				lca.Taxid(),
				lca.ScientificName(),
				lca.Rank())
		}

		if lca.Taxid() == 1 {
			break
		}
	}

	return index
}
