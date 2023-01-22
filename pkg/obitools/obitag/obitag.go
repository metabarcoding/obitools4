package obitag

import (
	"log"
	"math"
	"strconv"
	"strings"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/goutils"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obialign"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obikmer"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitax"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obifind"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obirefidx"
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

	o := goutils.ReverseIntOrder(cw)

	// mcw := 100000
	// for _, i := range o {
	// 	if cw[i] < mcw {
	// 		mcw = cw[i]
	// 	}
	// 	if cw[i] > mcw {
	// 		log.Panicln("wrong order")
	// 	}
	// }

	bests := obiseq.MakeBioSequenceSlice()
	bests = append(bests, references[o[0]])
	bestidxs := make([]int, 0)
	bestidxs = append(bestidxs, o[0])
	bestId := 0.0
	bestmatch := references[o[0]].Id()

	maxe := 0
	n := 0
	nf := 0

	for i, j := range o {
		ref := references[j]

		lmin, lmax := goutils.MinMaxInt(sequence.Len(), ref.Len())
		atMost := lmax - lmin + int(math.Ceil(float64(lmin-3-cw[j])/4.0)) - 2

		if i == 0 {
			maxe = goutils.MaxInt(sequence.Len(), ref.Len())
		}

		// log.Println(sequence.Id(),cw[j], maxe)
		if runExact || (atMost <= (maxe + 1)) {
			// if true {
			lcs, alilength := obialign.FastLCSScore(sequence, ref, maxe+1, &matrix)
			// fmt.Println(j, cw[j], lcs, alilength, alilength-lcs)
			// lcs, alilength := obialign.LCSScore(sequence, ref, maxe+1, matrix)
			n++
			if lcs == -1 {
				nf++
				// That aligment is worst than maxe, go to the next sequence
				continue
			}

			score := alilength - lcs
			if score < maxe {
				bests = bests[:0]
				bestidxs = bestidxs[:0]
				maxe = score
				bestId = float64(lcs) / float64(alilength)
				// log.Println(best.Id(), maxe, bestId)
			}

			if score == maxe {
				bests = append(bests, ref)
				bestidxs = append(bestidxs, j)
				id := float64(lcs) / float64(alilength)
				if id > bestId {
					bestId = id
					bestmatch = ref.Id()
				}
			}

		}

		if maxe == 0 {
			// We have found identity no need to continue to search
			break
		}
	}
	// log.Println("that's all falks", n, nf, maxe, bestId, bestidx)
	return bests, maxe, bestId, bestmatch, bestidxs
}

func Identify(sequence *obiseq.BioSequence,
	references obiseq.BioSequenceSlice,
	refcounts []*obikmer.Table4mer,
	taxo *obitax.Taxonomy,
	runExact bool) *obiseq.BioSequence {
	bests, differences, identity, bestmatch, seqidxs := FindClosests(sequence, references, refcounts, runExact)

	taxon := (*obitax.TaxNode)(nil)

	for i, best := range bests {
		idx := best.OBITagRefIndex()
		if idx == nil {
			// log.Fatalln("Need of indexing")
			idx = obirefidx.IndexSequence(seqidxs[i], references, taxo)
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
	taxo *obitax.Taxonomy,
	runExact bool) obiiter.SeqWorker {
	return func(sequence *obiseq.BioSequence) *obiseq.BioSequence {
		return Identify(sequence, references, refcounts, taxo, runExact)
	}
}

func AssignTaxonomy(iterator obiiter.IBioSequence) obiiter.IBioSequence {

	references := CLIRefDB()
	refcounts := make(
		[]*obikmer.Table4mer,
		len(references))

	buffer := make([]byte, 0, 1000)

	for i, seq := range references {
		refcounts[i] = obikmer.Count4Mer(seq, &buffer, nil)
	}

	taxo, error := obifind.CLILoadSelectedTaxonomy()

	if error != nil {
		log.Panicln(error)
	}

	worker := IdentifySeqWorker(references, refcounts, taxo, CLIRunExact())

	return iterator.Rebatch(17).MakeIWorker(worker, obioptions.CLIParallelWorkers(), 0).Speed("Annotated sequences").Rebatch(1000)
}
