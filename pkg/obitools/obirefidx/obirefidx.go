package obirefidx

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obifind"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	"github.com/schollz/progressbar/v3"
)

func IndexSequence(seqidx int,
	references obiseq.BioSequenceSlice,
	kmers *[]*obikmer.Table4mer,
	taxa *obitax.TaxonSet,
	taxo *obitax.Taxonomy) map[int]string {

	sequence := references[seqidx]

	var matrix []uint64

	lca := make(obitax.TaxonSet, len(references))
	tseq := (*taxa)[seqidx]

	for i, taxon := range *taxa {
		lca[i], _ = tseq.LCA(taxon)
	}

	cw := make([]int, len(references))
	sw := (*kmers)[seqidx]
	for i, ref := range *kmers {
		cw[i] = obikmer.Common4Mer(sw, ref)
	}

	ow := obiutils.Reverse(obiutils.IntOrder(cw), true)
	pseq, _ := tseq.Path()
	obiutils.Reverse(*pseq, true)
	// score := make([]int, len(references))
	mindiff := make([]int, len(*pseq))
	/*
		 	nseq := make([]int, len(*pseq))
			nali := make([]int, len(*pseq))
			nok := make([]int, len(*pseq))
			nfast := make([]int, len(*pseq))
			nfastok := make([]int, len(*pseq))
	*/

	lseq := sequence.Len()

	mini := -1
	wordmin := 0

	for i, ancestor := range *pseq {
		for _, order := range ow {
			if lca[order] == ancestor {
				// nseq[i]++
				if mini != -1 {
					wordmin = obiutils.MaxInt(sequence.Len(), references[order].Len()) - 3 - 4*mini
				}

				if cw[order] < wordmin {
					break
				}

				lcs, alilength := -1, -1
				errs := int(1e9)
				if mini != -1 && mini <= 1 {
					// nfast[i]++
					d, _, _, _ := obialign.D1Or0(sequence, references[order])
					if d >= 0 {
						errs = d
						// nfastok[i]++
					}
				} else {
					// nali[i]++
					lcs, alilength = obialign.FastLCSScore(sequence, references[order], mini, &matrix)
					if lcs >= 0 {
						// nok[i]++
						errs = alilength - lcs
					}

				}
				if mini == -1 || errs < mini {
					mini = errs
				}
			}
		}
		mindiff[i] = mini
	}

	obitag_index := make(map[int]string, len(*pseq))

	old := lseq
	for i, d := range mindiff {
		if d != -1 && d < old {
			current_taxid := (*pseq)[i]
			obitag_index[d] = fmt.Sprintf(
				"%d@%s@%s",
				current_taxid.Taxid(),
				current_taxid.ScientificName(),
				current_taxid.Rank())
			old = d
		}
	}

	/*
		 	log.Println(sequence.Id(), tseq.Taxid(), tseq.ScientificName(), tseq.Rank(), obitag_index)
			log.Println(sequence.Id(), tseq.Taxid(), tseq.ScientificName(), tseq.Rank(), nseq)
			log.Println(sequence.Id(), tseq.Taxid(), tseq.ScientificName(), tseq.Rank(), nfast)
			log.Println(sequence.Id(), tseq.Taxid(), tseq.ScientificName(), tseq.Rank(), nfastok)
			log.Println(sequence.Id(), tseq.Taxid(), tseq.ScientificName(), tseq.Rank(), nali)
			log.Println(sequence.Id(), tseq.Taxid(), tseq.ScientificName(), tseq.Rank(), nok)
	*/return obitag_index
}

func IndexReferenceDB(iterator obiiter.IBioSequence) obiiter.IBioSequence {

	log.Infoln("Loading database...")
	references := iterator.Load()
	log.Infof("Done. Database contains %d sequences", len(references))

	taxo, error := obifind.CLILoadSelectedTaxonomy()
	if error != nil {
		log.Panicln(error)
	}

	log.Infoln("Indexing sequence taxids...")

	taxa := make(
		obitax.TaxonSet,
		len(references))

	j := 0
	for i, seq := range references {
		taxon, err := taxo.Taxon(seq.Taxid())
		if err == nil {
			taxa[j] = taxon
			references[j] = references[i]
			j++
		}
	}

	if j < len(references) {
		if len(references)-j == 1 {
			log.Infoln("1 sequence has no valid taxid and has been discarded")
		} else {
			log.Infof("%d sequences have no valid taxid and has been discarded", len(references)-j)
		}

		references = references[0:j]
	} else {
		log.Infoln("Done.")
	}

	log.Infoln("Indexing database kmers...")
	refcounts := make(
		[]*obikmer.Table4mer,
		len(references))

	buffer := make([]byte, 0, 1000)

	for i, seq := range references {
		refcounts[i] = obikmer.Count4Mer(seq, &buffer, nil)
	}

	log.Info("done")

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
			limits <- [2]int{i, obiutils.MinInt(i+10, len(references))}
		}
		close(limits)
	}()

	f := func() {
		for l := range limits {
			sl := obiseq.MakeBioSequenceSlice()
			for i := l[0]; i < l[1]; i++ {
				idx := IndexSequence(i, references, &refcounts, &taxa, taxo)
				iref := references[i].Copy()
				iref.SetOBITagRefIndex(idx)
				sl = append(sl, iref)
				bar.Add(1)
			}
			indexed.Push(obiiter.MakeBioSequenceBatch(l[0]/10, sl))
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
