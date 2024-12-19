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
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	"github.com/schollz/progressbar/v3"
)

// IndexSequence processes a biological sequence and indexes it based on taxonomic information.
// It computes the least common ancestors (LCA) for the sequence and its references,
// evaluates common k-mers, and calculates differences in alignment scores.
//
// Parameters:
//   - seqidx: The index of the sequence to process within the references slice.
//   - references: A slice of biological sequences to compare against.
//   - kmers: A pointer to a slice of k-mer tables used for common k-mer calculations.
//   - taxa: A slice of taxonomic information corresponding to the sequences.
//   - taxo: A taxonomy object used for LCA calculations.
//
// Returns:
//
//	A map where the keys are integers representing alignment differences,
//	and the values are strings formatted as "Taxon@Rank" indicating the taxonomic
//	classification of the sequence based on the computed differences.
func IndexSequence(seqidx int,
	references obiseq.BioSequenceSlice,
	kmers *[]*obikmer.Table4mer,
	taxa *obitax.TaxonSlice,
	taxo *obitax.Taxonomy) map[int]string {

	// Retrieve the biological sequence at the specified index from the references slice.
	sequence := references[seqidx]
	seq_len := sequence.Len()
	// Get the taxon corresponding to the current sequence index.
	tseq := taxa.Taxon(seqidx)

	// Get the taxonomic path for the current sequence.
	pseq := tseq.Path()
	path_len := pseq.Len()

	// For each taxonomic ancestor in the path, a biosequence slice is created to store
	// the reference sequences having that ancestor as their LCA with the current sequence.
	refs := make(map[*obitax.TaxNode]*[]int, path_len)

	for i := 0; i < path_len; i++ {
		temp := make([]int, 0, 100)
		refs[pseq.Taxon(i).Node] = &temp
	}

	// log.Infof("%s length of path: %d", sequence.Id(), len(refs))

	n := taxa.Len()
	lcaCache := make(map[*obitax.TaxNode]*obitax.TaxNode, n)

	for i := 0; i < n; i++ {
		taxon := taxa.Taxon(i) // Get the taxon at index i.
		// Compute the LCA between the current taxon and the taxon of the sequence.
		node, ok := lcaCache[taxon.Node]
		if !ok {
			lca, err := tseq.LCA(taxon)
			if err != nil {
				// Log a fatal error if the LCA computation fails, including the taxon details.
				log.Fatalf("(%s,%s): %+v", tseq.String(), taxon.String(), err)
			}
			node = lca.Node
			lcaCache[taxon.Node] = node
		}

		// log.Infof("%s Processing taxon: %s x %s -> %s", sequence.Id(), tseq.String(), taxon.String(), node.String(taxo.Code()))

		// Append the current sequence to the LCA's reference sequence slice.
		*refs[node] = append(*refs[node], i)
	}

	closest := make([]int, path_len)

	closest[0] = 0

	// Initialize a matrix to store alignment scores
	var matrix []uint64

	// log.Warnf("%s : %s", sequence.Id(), pseq.String())
	for idx_path := 1; idx_path < path_len; idx_path++ {
		mini := -1
		seqidcs := refs[pseq.Taxon(idx_path).Node]

		ns := len(*seqidcs)

		if ns > 0 {

			shared := make([]int, ns)

			for j, is := range *seqidcs {
				shared[j] = obikmer.Common4Mer((*kmers)[seqidx], (*kmers)[is])
			}

			ow := obiutils.Reverse(obiutils.IntOrder(shared), true)

			for _, order := range ow {
				is := (*seqidcs)[order]
				suject := references[is]

				if mini != -1 {
					wordmin := max(seq_len, suject.Len()) - 3 - 4*mini

					// If the common k-mer count for the current order is less than the
					// minimum word length, break the loop.
					if shared[order] < wordmin {
						break
					}
				}

				// Initialize variables for Longest Common Subsequence (LCS) score and alignment length.
				lcs, alilength := -1, -1
				errs := int(1e9) // Initialize errors to a large number.

				// If mini is set and less than or equal to 1, perform a specific alignment.
				if mini == 0 || mini == 1 {
					// Perform a specific alignment and get the distance.
					d, _, _, _ := obialign.D1Or0(sequence, suject)
					if d >= 0 { // If the distance is valid (non-negative).
						errs = d // Update errors with the distance.
					}
				} else {
					// Perform a Fast LCS score calculation for the sequence and reference.
					lcs, alilength = obialign.FastLCSScore(sequence, suject, mini, &matrix)
					if lcs >= 0 { // If LCS score is valid (non-negative).
						errs = alilength - lcs // Calculate errors based on alignment length.
					}
				}

				// Update mini with the minimum errors found.
				if mini == -1 || errs < mini {
					mini = errs
				}

				if mini == 0 {
					// log.Warnf("%s: %s", sequence.Id(), sequence.String())
					// log.Warnf("%s: %s", suject.Id(), suject.String())
					break
				}
			}

			if mini == -1 {
				log.Fatalf("(%s,%s): No alignment found.", sequence.Id(), pseq.Taxon(idx_path).String())
			}

			closest[idx_path] = mini
			// insure than closest is strictly increasing
			for k := idx_path - 1; k >= 0 && mini < closest[k]; k-- {
				closest[k] = mini
				// log.Warnf("(%s,%s) Smaller alignment found than previous (%d,%d). Resetting closest.", sequence.Id(), pseq.Taxon(idx_path).String(), mini, closest[k])
			}
		} else {
			closest[idx_path] = seq_len
		}
	}

	obitag_index := make(map[int]string, pseq.Len())

	// log.Warnf("(%s,%s): %v", sequence.Id(), pseq.Taxon(0).String(), closest)
	for i, d := range closest {
		if i < (len(closest)-1) && d < closest[i+1] {
			current_taxon := pseq.Taxon(i)
			obitag_index[d] = fmt.Sprintf(
				"%s@%s",
				current_taxon.String(),
				current_taxon.Rank(),
			)
		}
	}

	/*
		 	log.Println(sequence.Id(), tseq.Taxid(), tseq.ScientificName(), tseq.Rank(), obitag_index)
			log.Println(sequence.Id(), tseq.Taxid(), tseq.ScientificName(), tseq.Rank(), nseq)
			log.Println(sequence.Id(), tseq.Taxid(), tseq.ScientificName(), tseq.Rank(), nfast)
			log.Println(sequence.Id(), tseq.Taxid(), tseq.ScientificName(), tseq.Rank(), nfastok)
			log.Println(sequence.Id(), tseq.Taxid(), tseq.ScientificName(), tseq.Rank(), nali)
			log.Println(sequence.Id(), tseq.Taxid(), tseq.ScientificName(), tseq.Rank(), nok)
	*/
	return obitag_index
}

func IndexReferenceDB(iterator obiiter.IBioSequence) obiiter.IBioSequence {

	log.Infoln("Loading database...")
	source, references := iterator.Load()
	log.Infof("Done. Database contains %d sequences", len(references))

	taxo, error := obioptions.CLILoadSelectedTaxonomy()
	if error != nil {
		log.Panicln(error)
	}

	log.Infoln("Indexing sequence taxids...")

	n := len(references)
	taxa := taxo.NewTaxonSlice(n, n)

	j := 0
	for i, seq := range references {
		taxon := seq.Taxon(taxo)
		if taxon != nil {
			taxa.Set(j, taxon)
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
			limits <- [2]int{i, min(i+10, len(references))}
		}
		close(limits)
	}()

	f := func() {
		for l := range limits {
			sl := obiseq.MakeBioSequenceSlice()
			for i := l[0]; i < l[1]; i++ {
				idx := IndexSequence(i, references, &refcounts, taxa, taxo)
				iref := references[i].Copy()
				iref.SetOBITagRefIndex(idx)
				sl = append(sl, iref)
				bar.Add(1)
			}
			indexed.Push(obiiter.MakeBioSequenceBatch(source, l[0]/10, sl))
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

	return indexed.Rebatch(obioptions.CLIBatchSize())
}
