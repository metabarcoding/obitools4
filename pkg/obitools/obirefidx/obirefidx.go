package obirefidx

import (
	"fmt"
	"log"
	"os"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/goutils"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obialign"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obikmer"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitax"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obifind"
	"github.com/schollz/progressbar/v3"
)

func IndexSequence(seqidx int,
	references obiseq.BioSequenceSlice,
	taxo *obitax.Taxonomy) map[int]string {

	sequence := references[seqidx]
	// matrix := obialign.NewFullLCSMatrix(nil,
	// 	sequence.Length(),
	// 	sequence.Length())

	var matrix []uint64

	score := make([]int, len(references))
	// t := 0
	// r := 0
	// w := 0
	for i, ref := range references {
		lcs, alilength := obialign.FastLCSScore(sequence, ref, -1, &matrix)
		score[i] = alilength - lcs
	}

	// log.Println("Redone : ",r,"/",t,"(",w,")")

	o := goutils.IntOrder(score)

	current_taxid, err := taxo.Taxon(references[o[0]].Taxid())
	current_score := score[o[0]]
	current_idx := o[0]

	if err != nil {
		log.Panicln(err)
	}

	obitag_index := make(map[int]string)

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
				obitag_index[score[current_idx]] = fmt.Sprintf(
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

	obitag_index[score[current_idx]] = fmt.Sprintf(
		"%d@%s@%s",
		current_taxid.Taxid(),
		current_taxid.ScientificName(),
		current_taxid.Rank())

	return obitag_index
}

func IndexReferenceDB(iterator obiiter.IBioSequence) obiiter.IBioSequence {

	references := iterator.Load()
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

	pbopt := make([]progressbar.Option, 0, 5)
	pbopt = append(pbopt,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(15),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetDescription("[Sequence Processing]"),
	)

	bar := progressbar.NewOptions(len(references), pbopt...)

	limits := make(chan [2]int)
	indexed := obiiter.MakeIBioSequence()
	go func() {
		for i := 0; i < len(references); i += 10 {
			limits <- [2]int{i, goutils.MinInt(i+10, len(references))}
		}
		close(limits)
	}()

	f := func() {
		for l := range limits {
			sl := obiseq.MakeBioSequenceSlice()
			for i := l[0]; i < l[1]; i++ {
				idx := IndexSequence(i, references, taxo)
				iref := references[i].Copy()
				iref.SetAttribute("obitag_ref_index", idx)
				sl = append(sl, iref)
			}
			indexed.Push(obiiter.MakeBioSequenceBatch(l[0]/10, sl))
			bar.Add(len(sl))
		}

		indexed.Done()
	}

	nworkers := obioptions.CLIParallelWorkers()
	indexed.Add(nworkers)

	go func() {
		indexed.WaitAndClose()
	}()

	for w := 0; w < nworkers; w++ {
		go f()
	}

	return indexed.Rebatch(1000)
}
