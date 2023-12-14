package obitax

import (
	"math"
	"strconv"
	"strings"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

func (t1 *TaxNode) LCA(t2 *TaxNode) (*TaxNode, error) {
	if t1 == nil {
		log.Fatalf("Try to get LCA of nil taxon")
	}

	if t2 == nil {
		log.Fatalf("Try to get LCA of nil taxon")
	}

	p1, err1 := t1.Path()

	if err1 != nil {
		return nil, err1
	}

	p2, err2 := t2.Path()

	if err2 != nil {
		return nil, err2
	}

	i1 := len(*p1) - 1
	i2 := len(*p2) - 1

	for i1 >= 0 && i2 >= 0 && (*p1)[i1].taxid == (*p2)[i2].taxid {
		i1--
		i2--
	}

	return (*p1)[i1+1], nil
}

func (taxonomy *Taxonomy) TaxonomicDistribution(sequence *obiseq.BioSequence) map[*TaxNode]int {
	taxids := sequence.StatsOn("taxid", "na")
	taxons := make(map[*TaxNode]int, len(taxids))

	for k, v := range taxids {
		taxid, _ := strconv.Atoi(k)

		t, et := taxonomy.Taxon(taxid)
		if et != nil {
			log.Panicf("Taxid %d not defined in taxonomy : %v", taxid, et)
		}
		taxons[t] = v
	}
	return taxons
}

func (taxonomy *Taxonomy) LCA(sequence *obiseq.BioSequence, threshold float64) (*TaxNode, float64, int) {
	taxons := taxonomy.TaxonomicDistribution(sequence)
	paths := make(map[*TaxNode]*TaxonSlice, len(taxons))
	answer := (*TaxNode)(nil)
	rans := 1.0
	granTotal := 0

	for t, w := range taxons {
		p, ep := t.Path()
		if ep != nil {
			log.Panicf("Taxonomic path cannot be retreived from Taxid %d : %v", t.Taxid(), ep)
		}

		obiutils.Reverse(*p, true)
		paths[t] = p
		answer = (*p)[0]
		granTotal += w
	}

	rmax := 1.0
	levels := make(map[*TaxNode]int, len(paths))
	taxonMax := answer

	for i := 0; rmax >= threshold; i++ {
		answer = taxonMax
		rans = rmax
		taxonMax = nil
		total := 0
		for taxon, weight := range taxons {
			path := paths[taxon]
			if len(*path) > i {
				levels[(*path)[i]] += weight
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
			if i < len(*path) {
				if (*path)[i] != taxonMax {
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
	return answer, rans, granTotal

}

func AddLCAWorker(taxonomy *Taxonomy, slot_name string, threshold float64) obiseq.SeqWorker {

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

	f := func(sequence *obiseq.BioSequence) *obiseq.BioSequence {
		lca, rans, _ := taxonomy.LCA(sequence, threshold)

		sequence.SetAttribute(slot_name, lca.Taxid())
		sequence.SetAttribute(lca_name, lca.ScientificName())
		sequence.SetAttribute(lca_error, math.Round((1-rans)*1000)/1000)

		return sequence
	}

	return f
}
