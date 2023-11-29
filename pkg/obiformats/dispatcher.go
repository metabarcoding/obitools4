package obiformats

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
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

				key := dispatcher.Classifier().Value(newflux)
				directory := ""
				if dispatcher.Classifier().Type == "DualAnnotationClassifier" {
					var keys [2]string
					err := json.Unmarshal([]byte(key), &keys)
					if err != nil {
						log.Fatalf("Error in parsing dispatch key %s", key)
					}
					key = keys[0]
					directory = keys[1]
				}

				name := fmt.Sprintf(prototypename, key)
				if opt.CompressedFile() && !strings.HasSuffix(name, ".gz") {
					name = name + ".gz"
				}

				if directory != "" {
					info, err := os.Stat(directory)
					switch {
					case !os.IsNotExist(err) && !info.IsDir():
						log.Fatalf("Cannot Create the directory %s", directory)
					case os.IsNotExist(err):
						os.Mkdir(directory, 0755)
					}

					name = filepath.Join(directory, name)
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
