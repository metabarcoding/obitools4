package obitag

import (
	"fmt"
	"log"
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
)

func IndexSequence(seqidx int,
	references obiseq.BioSequenceSlice,
	refcounts []*obikmer.Table4mer,
	taxo *obitax.Taxonomy) map[int]string {

	sequence := references[seqidx]
	matrix := obialign.NewLCSMatrix(nil,
		sequence.Length(),
		sequence.Length(),
		sequence.Length())

	score := make([]int, len(references))
	for i, ref := range references {
		maxe := goutils.MaxInt(sequence.Length(), ref.Length())
		mine := 0
		if refcounts != nil {
			mine, maxe = obikmer.Error4MerBounds(refcounts[seqidx], refcounts[i])
		}
		lcs, alilength := obialign.LCSScore(sequence, ref, (maxe+1)*2, matrix)

		if lcs < 0 {
			log.Print("Max error wrongly estimated", mine, maxe)
			log.Println(string(sequence.Sequence()))
			log.Fatalln(string(ref.Sequence()))

			maxe := goutils.MaxInt(sequence.Length(), ref.Length())
			lcs, alilength = obialign.LCSScore(sequence, ref, maxe, matrix)
		}
		score[i] = alilength - lcs
	}

	o := goutils.IntOrder(score)

	current_taxid, err := taxo.Taxon(references[o[0]].Taxid())
	current_score := score[o[0]]
	current_idx := o[0]

	if err != nil {
		log.Panicln(err)
	}

	ecotag_index := make(map[int]string)

	for _, idx := range o {
		new_taxid, err := taxo.Taxon(references[idx].Taxid())
		if err != nil {
			log.Panicln(err)
		}

		new_taxid, err = current_taxid.LCA(new_taxid)
		if err != nil {
			log.Panicln(err)
		}

		new_score := score[idx]

		if current_taxid.Taxid() != new_taxid.Taxid() {

			if new_score > current_score {
				ecotag_index[score[current_idx]] = fmt.Sprintf(
					"%d@%s@%s",
					current_taxid.Taxid(),
					current_taxid.ScientificName(),
					current_taxid.Rank())
				current_score = new_score
			}

			current_taxid = new_taxid
			current_idx = idx
		}
	}

	ecotag_index[score[current_idx]] = fmt.Sprintf(
		"%d@%s@%s",
		current_taxid.Taxid(),
		current_taxid.ScientificName(),
		current_taxid.Rank())

	sequence.SetAttribute("ecotag_ref_index", ecotag_index)

	return ecotag_index
}

func FindClosest(sequence *obiseq.BioSequence,
	references obiseq.BioSequenceSlice) (*obiseq.BioSequence, int, float64, int) {

	matrix := obialign.NewLCSMatrix(nil,
		sequence.Length(),
		sequence.Length(),
		sequence.Length())

	maxe := goutils.MaxInt(sequence.Length(), references[0].Length())
	best := references[0]
	bestidx := 0
	bestId := 0.0

	for i, ref := range references {
		lcs, alilength := obialign.LCSScore(sequence, ref, maxe, matrix)
		if lcs == -1 {
			// That aligment is worst than maxe, go to the next sequence
			continue
		}

		score := alilength - lcs
		if score < maxe {
			best = references[i]
			bestidx = i
			maxe = score
			bestId = float64(lcs) / float64(alilength)
			// log.Println(best.Id(), maxe, bestId)
		}

		if maxe == 0 {
			// We have found identity no need to continue to search
			break
		}
	}
	return best, maxe, bestId, bestidx
}

func Identify(sequence *obiseq.BioSequence,
	references obiseq.BioSequenceSlice,
	refcounts []*obikmer.Table4mer,
	taxo *obitax.Taxonomy) *obiseq.BioSequence {
	best, differences, identity, seqidx := FindClosest(sequence, references)

	idx := best.EcotagRefIndex()
	if idx == nil {
		idx = IndexSequence(seqidx, references, refcounts, taxo)
	}

	d := differences
	identification, ok := idx[d]
	for !ok && d >= 0 {
		identification, ok = idx[d]
		d--
	}

	parts := strings.Split(identification, "@")
	taxid, err := strconv.Atoi(parts[0])

	if err != nil {
		log.Panicln("Cannot extract taxid from :", identification)
	}

	sequence.SetTaxid(taxid)
	sequence.SetAttribute("scientific_name", parts[1])
	sequence.SetAttribute("ecotag_rank", parts[2])
	sequence.SetAttribute("ecotag_id", identity)
	sequence.SetAttribute("ecotag_difference", differences)
	sequence.SetAttribute("ecotag_match", best.Id())

	return sequence
}

func IdentifySeqWorker(references obiseq.BioSequenceSlice,
	refcounts []*obikmer.Table4mer,
	taxo *obitax.Taxonomy) obiiter.SeqWorker {
	return func(sequence *obiseq.BioSequence) *obiseq.BioSequence {
		return Identify(sequence, references, refcounts, taxo)
	}
}

func AssignTaxonomy(iterator obiiter.IBioSequenceBatch) obiiter.IBioSequenceBatch {

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

	worker := IdentifySeqWorker(references, refcounts, taxo)

	return iterator.Rebatch(10).MakeIWorker(worker, obioptions.CLIParallelWorkers(), 0).Rebatch(1000)
}
