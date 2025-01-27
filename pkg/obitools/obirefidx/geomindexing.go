package obirefidx

import (
	"fmt"
	"sort"
	"sync"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obistats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

func GeomIndexSesquence(seqidx int,
	references obiseq.BioSequenceSlice,
	taxa *obitax.TaxonSlice,
	taxo *obitax.Taxonomy) map[int]string {

	sequence := references[seqidx]
	location := sequence.GetCoordinate()

	if location == nil {
		log.Fatalf("Sequence %s does not have a coordinate", sequence.Id())
	}

	seq_dist := make([]int, len(references))

	var wg sync.WaitGroup

	iseq_channel := make(chan int)

	for k := 0; k < obidefault.ParallelWorkers(); k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range iseq_channel {
				ref := references[i]
				reflocation := ref.GetCoordinate()
				if reflocation == nil {
					log.Fatalf("Sequence %s does not have a coordinate", ref.Id())
				}

				seq_dist[i] = obistats.SquareDist(location, reflocation)
			}
		}()
	}

	for i := range references {
		iseq_channel <- i
	}

	close(iseq_channel)
	wg.Wait()

	order := obiutils.Order(sort.IntSlice(seq_dist))

	lca := taxa.Taxon(seqidx)

	index := make(map[int]string)
	index[0] = fmt.Sprintf(
		"%s@%s",
		lca.String(),
		lca.Rank(),
	)

	for _, o := range order {
		new_lca, _ := lca.LCA(taxa.Taxon(o))
		if new_lca.SameAs(lca) {
			lca = new_lca
			index[int(seq_dist[o])] = lca.String()

			if lca.IsRoot() {
				break
			}
		}
	}

	return index
}
