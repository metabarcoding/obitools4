package obirefidx

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obialign"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obikmer"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitax"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obifind"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiutils"
	"github.com/schollz/progressbar/v3"
)

func IndexSequence(seqidx int,
	references obiseq.BioSequenceSlice,
	kmers *[]*obikmer.Table4mer,
	taxa *obitax.TaxonSet,
	taxo *obitax.Taxonomy) map[int]string {

	sequence := references[seqidx]
	// matrix := obialign.NewFullLCSMatrix(nil,
	// 	sequence.Length(),
	// 	sequence.Length())

	var matrix []uint64

	lca := make(obitax.TaxonSet, len(references))
	tref := (*taxa)[seqidx]

	for i, taxon := range (*taxa) {
		lca[i],_ = tref.LCA(taxon)
	}

	cw := make([]int, len(references))
	sw := (*kmers)[seqidx]
	for i, ref := range *kmers {
		cw[i] = obikmer.Common4Mer(sw,ref)
	}

	ow := obiutils.Reverse(obiutils.IntOrder(cw),true)
	pref,_ := tref.Path()
	obiutils.Reverse(*pref,true)
	// score := make([]int, len(references))
	mindiff := make([]int, len(*pref))


	for i,ancestor := range *pref {
		mini := -1
		for _,order := range ow {
			if lca[order] == ancestor {
				lcs, alilength := obialign.FastLCSScore(sequence, references[order], mini, &matrix)
				if lcs >= 0 {
					errs := alilength - lcs
					if mini== -1 || errs < mini {
						mini = errs
					}	
				}
			}
		}
		if mini != -1 {
			mindiff[i] = mini
		} else {
			mindiff[i] = 1e6
		}
	} 

	obitag_index := make(map[int]string, len(*pref)) 

	old := sequence.Len()
	for i,d := range mindiff {
		if d < old {
			current_taxid :=(*pref)[i]
			obitag_index[d] = fmt.Sprintf(
				"%d@%s@%s",
				current_taxid.Taxid(),
				current_taxid.ScientificName(),
				current_taxid.Rank())
				old = d
		}
	}

	/* // t := 0
	// r := 0
	// w := 0
	for i, ref := range references {
		lcs, alilength := obialign.FastLCSScore(sequence, ref, -1, &matrix)
		score[i] = alilength - lcs
	}

	// log.Println("Redone : ",r,"/",t,"(",w,")")

	o := obiutils.IntOrder(score)

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
 */
	//log.Println(obitag_index)
	return obitag_index
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
				iref.SetAttribute("obitag_ref_index", idx)
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
