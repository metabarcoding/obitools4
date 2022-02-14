package obiuniq

import (
	"log"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obichunk"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func Unique(sequences obiseq.IBioSequenceBatch) obiseq.IBioSequenceBatch {

	newIter, err := obichunk.ISequenceChunk(sequences, 100, 2)

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
