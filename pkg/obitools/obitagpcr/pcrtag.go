package obitagpcr

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obimultiplex"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obipairing"
)

// IPCRTagPESequencesBatch performs paired-end sequence demultiplexing and tagging.
// It takes an iterator of paired-end biosequences, and various parameters to control
// the demultiplexing and tagging process. It returns a new iterator with the
// demultiplexed and tagged sequences.
//
// The function first checks if the input iterator contains paired-end sequences.
// It then creates a new iterator to hold the processed sequences, and launches
// a number of worker goroutines to process the input sequences in parallel.
//
// For each sequence pair, the function attempts to assemble the paired-end
// sequences into a consensus sequence, and then extracts the barcodes and
// other metadata from the consensus. The extracted information is then
// added as annotations to the original sequence pair.
//
// If the CLI option to reorient the sequences is enabled, the function will
// reverse-complement the sequences as needed to ensure the forward and reverse
// tags are correctly oriented.
//
// If the CLI option to conserve unassigned sequences is disabled, the function
// will filter out any sequences that could not be demultiplexed.
//
// If the CLI option to save unassigned sequences is enabled, the function will
// write those sequences to a separate file.
func IPCRTagPESequencesBatch(iterator obiiter.IBioSequence,
	gap, scale float64, delta, minOverlap int,
	minIdentity float64, fastAlign, fastScoreRel,
	withStats bool) obiiter.IBioSequence {

	if !iterator.IsPaired() {
		log.Fatalln("Sequence data must be paired")
	}

	nworkers := obidefault.ParallelWorkers()
	ngsfilter, err := obimultiplex.CLINGSFIlter()

	if err != nil {
		log.Fatalf("%v", err)
	}

	if obimultiplex.CLIAllowsIndel() {
		ngsfilter.SetAllowsIndels(true)
	}

	if obimultiplex.CLIAllowedMismatch() > 0 {
		ngsfilter.SetAllowedMismatches(obimultiplex.CLIAllowedMismatch())
	}

	ngsfilter.Compile2()

	newIter := obiiter.MakeIBioSequence()
	newIter.MarkAsPaired()

	f := func(iterator obiiter.IBioSequence) {
		arena := obialign.MakePEAlignArena(150, 150)
		shifts := make(map[int]int)

		for iterator.Next() {
			batch := iterator.Get()
			for i, A := range batch.Slice() {
				B := A.PairedWith()
				consensus := obipairing.AssemblePESequences(
					A.Copy(), B.ReverseComplement(false),
					gap, scale,
					delta, minOverlap, minIdentity, withStats, true,
					fastAlign, fastScoreRel, arena, &shifts,
				)

				barcodes, err := ngsfilter.ExtractMultiBarcode(consensus)

				if len(barcodes) == 1 && !barcodes[0].HasAttribute("obimultiplex_error") && err == nil {
					consensus = barcodes[0]

					annot := consensus.Annotations()
					direction := annot["obimultiplex_direction"].(string)

					forward_match := annot["obimultiplex_forward_match"].(string)
					forward_mismatches := annot["obimultiplex_forward_error"].(int)

					reverse_match := annot["obimultiplex_reverse_match"].(string)
					reverse_mismatches := annot["obimultiplex_reverse_error"].(int)

					sample := annot["sample"].(string)
					experiment := annot["experiment"].(string)

					aanot := A.Annotations()
					banot := B.Annotations()

					if value, ok := annot["obimultiplex_forward_tag"]; ok {
						forward_tag := value.(string)
						aanot["obimultiplex_forward_tag"] = forward_tag
						banot["obimultiplex_forward_tag"] = forward_tag
					}

					if value, ok := annot["obimultiplex_reverse_tag"]; ok {
						reverse_tag := value.(string)
						aanot["obimultiplex_reverse_tag"] = reverse_tag
						banot["obimultiplex_reverse_tag"] = reverse_tag
					}

					aanot["obimultiplex_direction"] = direction

					aanot["obimultiplex_forward_match"] = forward_match
					aanot["obimultiplex_forward_mismatches"] = forward_mismatches

					aanot["obimultiplex_reverse_match"] = reverse_match
					aanot["obimultiplex_reverse_mismatches"] = reverse_mismatches

					aanot["sample"] = sample
					aanot["experiment"] = experiment

					banot["obimultiplex_direction"] = direction

					banot["obimultiplex_forward_match"] = forward_match
					banot["obimultiplex_forward_mismatches"] = forward_mismatches

					banot["obimultiplex_reverse_match"] = reverse_match
					banot["obimultiplex_reverse_mismatches"] = reverse_mismatches

					banot["sample"] = sample
					banot["experiment"] = experiment

					if CLIReorientate() && direction == "reverse" {
						B.PairTo(A)
						batch.Slice()[i] = B
					}
				} else {
					demultiplex_error := "Cannot demultiplex"
					if len(barcodes) > 0 {
						var ok bool
						consensus = barcodes[0]
						demultiplex_error, ok = consensus.GetStringAttribute("obimultiplex_error")
						if !ok {
							demultiplex_error = "Cannot demultiplex"
						}
					}
					A.Annotations()["obimultiplex_error"] = demultiplex_error
					B.Annotations()["obimultiplex_error"] = demultiplex_error

				}

				// log.Println("@@ ",wid," Error : ",err)
				// log.Println("@@ ",wid," SEQA : ",*A)
				// log.Println("@@ ",wid," SEQB : ",*B)
				// log.Println("@@ ",wid," consensus : ",*consensus)

			}
			newIter.Push(batch)
		}
		newIter.Done()
	}

	log.Printf("Start of the sequence Pairing using %d workers\n", nworkers)

	newIter.Add(nworkers)
	for i := 1; i < nworkers; i++ {
		go f(iterator.Split())
	}
	go f(iterator)

	go func() {
		newIter.WaitAndClose()
		log.Printf("End of the sequence PCR Taging")
	}()

	iout := newIter

	if !obimultiplex.CLIConservedErrors() {
		log.Println("Discards unassigned sequences")
		iout = iout.FilterOn(obiseq.HasAttribute("obimultiplex_error").Not(), obidefault.BatchSize())
	}

	var unidentified obiiter.IBioSequence
	if obimultiplex.CLIUnidentifiedFileName() != "" {
		log.Printf("Unassigned sequences saved in file: %s\n", obimultiplex.CLIUnidentifiedFileName())
		unidentified, iout = iout.DivideOn(obiseq.HasAttribute("obimultiplex_error"),
			obidefault.BatchSize())

		go func() {
			_, err := obiconvert.CLIWriteBioSequences(unidentified,
				true,
				obimultiplex.CLIUnidentifiedFileName())

			if err != nil {
				log.Fatalf("%v", err)
			}
		}()

	}
	log.Printf("Sequence demultiplexing using %d workers\n", nworkers)

	return iout

}
