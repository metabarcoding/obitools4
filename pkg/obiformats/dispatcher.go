package obiformats

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/goccy/go-json"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
)

// SequenceBatchWriterToFile is a function type that defines a method for writing
// a batch of biosequences to a specified file. It takes an iterator of biosequences,
// a filename, and optional configuration options, and returns an iterator of biosequences
// along with any error encountered during the writing process.
//
// Parameters:
// - iterator: An iterator of biosequences to be written to the file.
// - filename: The name of the file where the sequences will be written.
// - options: Optional configuration options for the writing process.
//
// Returns:
// An iterator of biosequences that may have been modified during the writing process
// and an error if the writing operation fails.
type SequenceBatchWriterToFile func(iterator obiiter.IBioSequence,
	filename string,
	options ...WithOption) (obiiter.IBioSequence, error)

// WriterDispatcher manages the writing of data to files based on a given
// prototype name and a dispatcher for distributing the sequences. It
// processes incoming data from the dispatcher in separate goroutines,
// formatting and writing the data to files as specified.
//
// Parameters:
//   - prototypename: A string that serves as a template for naming the output files.
//   - dispatcher: An instance of IDistribute that provides the data to be written
//     and manages the distribution of sequences.
//   - formater: A function of type SequenceBatchWriterToFile that formats and writes
//     the sequences to the specified file.
//   - options: Optional configuration options for the writing process.
//
// The function operates asynchronously, launching goroutines for each new data
// channel received from the dispatcher. It ensures that directories are created
// as needed and handles errors during the writing process. The function blocks
// until all writing jobs are completed.
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
					log.Fatalf("Cannot retrieve the new channel: %v", err)
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
						log.Fatalf("Cannot create the directory %s", directory)
					case os.IsNotExist(err):
						os.Mkdir(directory, 0755)
					}

					name = filepath.Join(directory, name)
				}

				out, err := formater(data,
					name,
					options...)

				if err != nil {
					log.Fatalf("Cannot open the output file for key %s",
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
