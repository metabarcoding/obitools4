package obitag

import (
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obialign"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obikmer"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitax"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obifind"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obirefidx"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiutils"
)

func FindClosests(sequence *obiseq.BioSequence,
	references obiseq.BioSequenceSlice,
	refcounts []*obikmer.Table4mer,
	runExact bool) (obiseq.BioSequenceSlice, int, float64, string, []int) {

	var matrix []uint64

	seqwords := obikmer.Count4Mer(sequence, nil, nil)
	cw := make([]int, len(refcounts))

	for i, ref := range refcounts {
		cw[i] = obikmer.Common4Mer(seqwords, ref)
	}

	o := obiutils.Reverse(obiutils.IntOrder(cw), true)

	bests := obiseq.MakeBioSequenceSlice()
//	bests = append(bests, references[o[0]])
	bestidxs := make([]int, 0)
//	bestidxs = append(bestidxs, o[0])
	bestId := 0.0
	bestmatch := references[o[0]].Id()

	maxe := -1
	wordmin := 0

	for _, order := range o {
		ref := references[order]

		if maxe != -1 {
			wordmin = obiutils.MaxInt(sequence.Len(), ref.Len()) - 4*maxe
		}

		lcs, alilength := -1, -1
		score := int(1e9)
		if maxe == 0 || maxe == 1 {
			d, _, _, _ := obialign.D1Or0(sequence, references[order])
			if d >= 0 {
				score = d
				alilength = obiutils.MaxInt(sequence.Len(), ref.Len()) 
				lcs = alilength - score
			}
		} else {
			if cw[order] >= wordmin {
				lcs, alilength = obialign.FastLCSScore(sequence, references[order], maxe, &matrix)
				if lcs >= 0 {
					score = alilength - lcs
				}
			}
		}

		if maxe == -1 || score < maxe {
			bests = bests[:0]
			bestidxs = bestidxs[:0]
			maxe = score
			bestId = float64(lcs) / float64(alilength)
		    // log.Println(ref.Id(), maxe, bestId,bestidxs)
		}

		if score == maxe {
			bests = append(bests, ref)
			bestidxs = append(bestidxs, order)
			id := float64(lcs) / float64(alilength)
			if id > bestId {
				bestId = id
				bestmatch = ref.Id()
			}
		    // log.Println(ref.Id(), maxe, bestId,bestidxs)
		}

	}

    //log.Println("that's all falks",  maxe, bestId, bestidxs)
	return bests, maxe, bestId, bestmatch, bestidxs
}

func Identify(sequence *obiseq.BioSequence,
	references obiseq.BioSequenceSlice,
	refcounts []*obikmer.Table4mer,
	taxa obitax.TaxonSet,
	taxo *obitax.Taxonomy,
	runExact bool) *obiseq.BioSequence {
	bests, differences, identity, bestmatch, seqidxs := FindClosests(sequence, references, refcounts, runExact)

	taxon := (*obitax.TaxNode)(nil)

	for i, best := range bests {
		idx := best.OBITagRefIndex()
		if idx == nil {
			// log.Fatalln("Need of indexing")
			idx = obirefidx.IndexSequence(seqidxs[i], references, &refcounts, &taxa, taxo)
		}

		d := differences
		identification, ok := idx[d]
		for !ok && d >= 0 {
			identification, ok = idx[d]
			d--
		}

		parts := strings.Split(identification, "@")
		match_taxid, err := strconv.Atoi(parts[0])

		if err != nil {
			log.Panicln("Cannot extract taxid from :", identification)
		}

		match_taxon, err := taxo.Taxon(match_taxid)

		if err != nil {
			log.Panicln("Cannot find taxon corresponding to taxid :", match_taxid)
		}

		if taxon != nil {
			taxon, _ = taxon.LCA(match_taxon)
		} else {
			taxon = match_taxon
		}

	}

	sequence.SetTaxid(taxon.Taxid())
	sequence.SetAttribute("scientific_name", taxon.ScientificName())
	sequence.SetAttribute("obitag_rank", taxon.Rank())
	sequence.SetAttribute("obitag_bestid", identity)
	sequence.SetAttribute("obitag_difference", differences)
	sequence.SetAttribute("obitag_bestmatch", bestmatch)
	sequence.SetAttribute("obitag_match_count", len(bests))

	return sequence
}

func IdentifySeqWorker(references obiseq.BioSequenceSlice,
	refcounts []*obikmer.Table4mer,
	taxa obitax.TaxonSet,
	taxo *obitax.Taxonomy,
	runExact bool) obiseq.SeqWorker {
	return func(sequence *obiseq.BioSequence) *obiseq.BioSequence {
		return Identify(sequence, references, refcounts, taxa, taxo, runExact)
	}
}

func AssignTaxonomy(iterator obiiter.IBioSequence) obiiter.IBioSequence {

	taxo, error := obifind.CLILoadSelectedTaxonomy()

	if error != nil {
		log.Panicln(error)
	}

	references := CLIRefDB()
	refcounts := make(
		[]*obikmer.Table4mer,
		len(references))

	taxa := make(obitax.TaxonSet,
		len(references))

	buffer := make([]byte, 0, 1000)

	for i, seq := range references {
		refcounts[i] = obikmer.Count4Mer(seq, &buffer, nil)
		taxa[i], _ = taxo.Taxon(seq.Taxid())
	}

	worker := IdentifySeqWorker(references, refcounts, taxa, taxo, CLIRunExact())

	return iterator.Rebatch(17).MakeIWorker(worker, obioptions.CLIParallelWorkers(), 0).Rebatch(1000)
}
