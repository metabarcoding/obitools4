package obiuniq

import (
	"log"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obichunk"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func Unique(sequences obiseq.IBioSequenceBatch) obiseq.IBioSequenceBatch {

	classifier := obiseq.HashClassifier(CLINumberOfChunks())
	var newIter obiseq.IBioSequenceBatch
	var err error

	if CLIUniqueInMemory() {
		log.Printf("Running dereplication in memory on %d chunks", CLINumberOfChunks())
		newIter, err = obichunk.ISequenceChunk(sequences, classifier, 2)
	} else {
		log.Printf("Running dereplication on disk with %d chunks", CLINumberOfChunks())
		newIter, err = obichunk.ISequenceChunkOnDisk(sequences, classifier, 2)
	}

	if err != nil {
		log.Fatalf("error in spliting the dataset : %v", err)
	}

	statsOn := CLIStatsOn()
	keys := CLIKeys()
	parallelWorkers := obioptions.CLIParallelWorkers()
	buffSize := obioptions.CLIBufferSize()

	newIter = newIter.MakeISliceWorker(obiseq.UniqueSliceWorker(statsOn, keys...),
		parallelWorkers, buffSize)

	return newIter
}
