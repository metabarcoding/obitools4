package obiformats

import (
	"fmt"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
)

type SequenceBatchWriterToFile func(iterator obiiter.IBioSequence,
	filename string,
	options ...WithOption) (obiiter.IBioSequence, error)

func WriterDispatcher(prototypename string,
	dispatcher obiiter.IDistribute,
	formater SequenceBatchWriterToFile,
	options ...WithOption) {

	jobDone := sync.WaitGroup{}
	jobDone.Add(1)

	go func() {
		opt := MakeOptions(options)
		for newflux := range dispatcher.News() {
			jobDone.Add(1)
			go func(newflux int) {
				data, err := dispatcher.Outputs(newflux)

				if err != nil {
					log.Fatalf("Cannot retreive the new chanel : %v", err)
				}

				name:=fmt.Sprintf(prototypename, dispatcher.Classifier().Value(newflux))
				if opt.CompressedFile() && ! strings.HasSuffix(name,".gz") {
					name = name + ".gz"
				}
				
				out, err := formater(data,
					name,
					options...)

				if err != nil {
					log.Fatalf("cannot open the output file for key %s",
						dispatcher.Classifier().Value(newflux))
				}

				out.Recycle()
				jobDone.Done()
			}(newflux)
		}
		jobDone.Done()
	}()

	jobDone.Wait()
}
