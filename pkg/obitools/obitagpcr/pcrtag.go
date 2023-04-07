package obitagpcr

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obialign"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obimultiplex"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obipairing"
)

func IPCRTagPESequencesBatch(iterator obiiter.IBioSequence,
	gap float64, delta, minOverlap int,
	minIdentity float64,
	withStats bool) obiiter.IBioSequence {

	if !iterator.IsPaired() {
		log.Fatalln("Sequence data must be paired")
	}

	nworkers := obioptions.CLIParallelWorkers()

	ngsfilter, err := obimultiplex.CLINGSFIlter()
	if err != nil {
		log.Fatalf("%v", err)
	}

	ngsfilter.Compile(obimultiplex.CLIAllowedMismatch(),
		obimultiplex.CLIAllowsIndel())

	newIter := obiiter.MakeIBioSequence()
	newIter.MarkAsPaired()
	newIter.Add(nworkers)

	go func() {
		newIter.WaitAndClose()
		log.Printf("End of the sequence PCR Taging")
	}()

	f := func(iterator obiiter.IBioSequence, wid int) {
		arena := obialign.MakePEAlignArena(150, 150)
		var err error

		for iterator.Next() {
			batch := iterator.Get()
			for i, A := range batch.Slice() {
				B := A.PairedWith()
				consensus := obipairing.AssemblePESequences(
					A.Copy(), B.ReverseComplement(false),
					gap, delta, minOverlap, minIdentity, withStats, true, arena,
				)

				consensus, err = ngsfilter.ExtractBarcode(consensus, true)

				// log.Println("@@ ",wid," Error : ",err)
				// log.Println("@@ ",wid," SEQA : ",*A)
				// log.Println("@@ ",wid," SEQB : ",*B)
				// log.Println("@@ ",wid," consensus : ",*consensus)

				if err == nil {
					annot := consensus.Annotations()
					direction := annot["direction"].(string)

					forward_match := annot["forward_match"].(string)
					forward_mismatches := annot["forward_mismatches"].(int)
					forward_tag := annot["forward_tag"].(string)

					reverse_match := annot["reverse_match"].(string)
					reverse_mismatches := annot["reverse_mismatches"].(int)
					reverse_tag := annot["reverse_tag"].(string)

					sample := annot["sample"].(string)
					experiment := annot["experiment"].(string)

					aanot := A.Annotations()
					aanot["direction"] = direction

					aanot["forward_match"] = forward_match
					aanot["forward_mismatches"] = forward_mismatches
					aanot["forward_tag"] = forward_tag

					aanot["reverse_match"] = reverse_match
					aanot["reverse_mismatches"] = reverse_mismatches
					aanot["reverse_tag"] = reverse_tag

					aanot["sample"] = sample
					aanot["experiment"] = experiment

					banot := B.Annotations()
					banot["direction"] = direction

					banot["forward_match"] = forward_match
					banot["forward_mismatches"] = forward_mismatches
					banot["forward_tag"] = forward_tag

					banot["reverse_match"] = reverse_match
					banot["reverse_mismatches"] = reverse_mismatches
					banot["reverse_tag"] = reverse_tag

					banot["sample"] = sample
					banot["experiment"] = experiment

					if CLIReorientate() && direction == "reverse" {
						B.ReverseComplement(true)
						A.ReverseComplement(true)
						B.PairTo(A)
						batch.Slice()[i] = B
					}
				} else {
					demultiplex_error := consensus.Annotations()["demultiplex_error"]
					if demultiplex_error!=nil {
						A.Annotations()["demultiplex_error"] = demultiplex_error.(string)
						B.Annotations()["demultiplex_error"] = demultiplex_error.(string)
					} else {
						log.Panicln("@@ ",wid,"Error : ",err,*consensus)
					}
				}
			}
			newIter.Push(batch)
		}
		newIter.Done()
	}

	log.Printf("Start of the sequence Pairing using %d workers\n", nworkers)

	for i := 1; i < nworkers; i++ {
		go f(iterator.Split(), i)
	}
	go f(iterator, 0)

	iout := newIter

	if !obimultiplex.CLIConservedErrors() {
		log.Println("Discards unassigned sequences")
		iout = iout.Rebatch(obioptions.CLIBatchSize())
	}

	var unidentified obiiter.IBioSequence
	if obimultiplex.CLIUnidentifiedFileName() != "" {
		log.Printf("Unassigned sequences saved in file: %s\n", obimultiplex.CLIUnidentifiedFileName())
		unidentified, iout = iout.DivideOn(obiseq.HasAttribute("demultiplex_error"),
			obioptions.CLIBatchSize())

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

// if match.IsDirect {
// 	annot["direction"] = "direct"
// } else {
// 	annot["direction"] = "reverse"
// }

// if match.ForwardMatch != "" {
// 	annot["forward_match"] = match.ForwardMatch
// 	annot["forward_mismatches"] = match.ForwardMismatches
// 	annot["forward_tag"] = match.ForwardTag
// }

// if match.ReverseMatch != "" {
// 	annot["reverse_match"] = match.ReverseMatch
// 	annot["reverse_mismatches"] = match.ReverseMismatches
// 	annot["reverse_tag"] = match.ReverseTag
// }

// if match.Error == nil {
// 	if match.Pcr != nil {
// 		annot["sample"] = match.Pcr.Sample
// 		annot["experiment"] = match.Pcr.Experiment
// 		for k, val := range match.Pcr.Annotations {
// 			annot[k] = val
// 		}
// 	} else {
// 		annot["demultiplex_error"]
