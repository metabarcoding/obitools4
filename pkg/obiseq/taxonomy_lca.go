package obiseq

import (
	"math"
	"strings"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	log "github.com/sirupsen/logrus"
)

func (sequence *BioSequence) TaxonomicDistribution(taxonomy *obitax.Taxonomy) map[*obitax.TaxNode]int {
	taxids := sequence.StatsOn(MakeStatsOnDescription("taxid"), "na")
	taxons := make(map[*obitax.TaxNode]int, taxids.Len())

	taxonomy = taxonomy.OrDefault(true)

	taxids.RLock()
	defer taxids.RUnlock()
	for taxid, v := range taxids.Map() {
		t, isAlias, err := taxonomy.Taxon(taxid)
		if err != nil {
			log.Fatalf(
				"On sequence %s taxid %s is not defined in taxonomy: %s (%v)",
				sequence.Id(),
				taxid,
				taxonomy.Name(),
				err,
			)
		}

		if isAlias && obidefault.FailOnTaxonomy() {
			log.Fatalf("On sequence %s taxid %s is an alias on %s",
				sequence.Id(), taxid, t.String())
		}
		taxons[t.Node] = v
	}
	return taxons
}

func (sequence *BioSequence) LCA(taxonomy *obitax.Taxonomy, threshold float64) (*obitax.Taxon, float64, int) {

	taxonomy = taxonomy.OrDefault(true)

	taxons := sequence.TaxonomicDistribution(taxonomy)
	paths := make(map[*obitax.TaxNode]*obitax.TaxonSlice, len(taxons))
	answer := (*obitax.TaxNode)(nil)
	rans := 1.0
	granTotal := 0

	for t, w := range taxons {
		taxon := &obitax.Taxon{Taxonomy: taxonomy, Node: t}
		p := taxon.Path()

		if p == nil {
			log.Panicf("Sequence %s: taxonomic path cannot be retreived from Taxid : %s", sequence.Id(), taxon.String())
		}

		p.Reverse(true)
		paths[t] = p
		answer = p.Get(0)
		granTotal += w
	}

	rmax := 1.0
	levels := make(map[*obitax.TaxNode]int, len(paths))
	taxonMax := answer

	for i := 0; rmax >= threshold; i++ {
		answer = taxonMax
		rans = rmax
		taxonMax = nil
		total := 0
		for taxon, weight := range taxons {
			path := paths[taxon]
			if path.Len() > i {
				levels[path.Get(i)] += weight
			}
			total += weight
		}
		weighMax := 0
		for taxon, weight := range levels {
			if weight > weighMax {
				weighMax = weight
				taxonMax = taxon
			}
		}

		if total > 0 {
			rmax *= float64(weighMax) / float64(total)
		} else {
			rmax = 0.0
		}

		for taxon := range levels {
			delete(levels, taxon)
		}
		for taxon := range taxons {
			path := paths[taxon]
			if i < path.Len() {
				if path.Get(i) != taxonMax {
					delete(paths, taxon)
					delete(taxons, taxon)
				}
			}
		}
		// if taxonMax != nil {
		// 	log.Println("@@@>", i, taxonMax.ScientificName(), taxonMax.Taxid(), rans, weighMax, total, rmax)
		// } else {
		// 	log.Println("@@@>", "--", 0, rmax)
		// }
	}
	// log.Println("###>", answer.ScientificName(), answer.Taxid(), rans)
	// log.Print("========================================")
	return &obitax.Taxon{Taxonomy: taxonomy, Node: answer}, rans, granTotal

}

func AddLCAWorker(taxonomy *obitax.Taxonomy, slot_name string, threshold float64) SeqWorker {

	taxonomy = taxonomy.OrDefault(true)

	if !strings.HasSuffix(slot_name, "taxid") {
		slot_name = slot_name + "_taxid"
	}

	lca_error := strings.Replace(slot_name, "taxid", "error", 1)
	if lca_error == "error" {
		lca_error = "lca_error"
	}

	lca_name := strings.Replace(slot_name, "taxid", "name", 1)
	if lca_name == "name" {
		lca_name = "scientific_name"
	}

	f := func(sequence *BioSequence) (BioSequenceSlice, error) {
		lca, rans, _ := sequence.LCA(taxonomy, threshold)

		sequence.SetAttribute(slot_name, lca.String())
		sequence.SetAttribute(lca_name, lca.ScientificName())
		sequence.SetAttribute(lca_error, math.Round((1-rans)*1000)/1000)
		return BioSequenceSlice{sequence}, nil
	}

	return f
}
