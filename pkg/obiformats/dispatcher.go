package obiformats

import (
	"fmt"
	"log"
	"sync"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

type SequenceBatchWriterToFile func(iterator obiseq.IBioSequenceBatch,
	filename string,
	options ...WithOption) (obiseq.IBioSequenceBatch, error)

func WriterDispatcher(prototypename string,
	dispatcher obiseq.IDistribute,
	formater SequenceBatchWriterToFile,
	options ...WithOption) {

	jobDone := sync.WaitGroup{}
	jobDone.Add(1)

	go func() {
		n := 0
		for newflux := range dispatcher.News() {
			data, _ := dispatcher.Outputs(newflux)
			out, err := formater(data,
				fmt.Sprintf(prototypename, newflux),
				options...)
			if err != nil {
				log.Fatalf("cannot open the output file for key %s", newflux)
			}

			n++

			if n > 1 {
				jobDone.Add(1)
			}

			go func() {
				out.Recycle()
				jobDone.Done()
			}()
		}
	}()

	jobDone.Wait()
}
