package obichunk

import (
	"io/fs"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiformats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// tempDir creates a temporary directory with a prefix "obiseq_chunks_"
// in the system's temporary directory. It returns the path of the
// created directory and any error encountered during the creation process.
//
// If the directory creation is successful, the path to the new
// temporary directory is returned. If there is an error, it returns
// an empty string and the error encountered.
func tempDir() (string, error) {
	dir, err := os.MkdirTemp(os.TempDir(), "obiseq_chunks_")
	if err != nil {
		return "", err
	}
	return dir, nil
}

// find searches for files with a specific extension in the given root directory
// and its subdirectories. It returns a slice of strings containing the paths
// of the found files.
//
// Parameters:
// - root: The root directory to start the search from.
// - ext: The file extension to look for (including the leading dot, e.g., ".txt").
//
// Returns:
// A slice of strings containing the paths of files that match the specified
// extension. If no files are found, an empty slice is returned. Any errors
// encountered during the directory traversal will be returned as part of the
// WalkDir function's error handling.
func find(root, ext string) []string {
	var a []string
	filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ext {
			a = append(a, s)
		}
		return nil
	})
	return a
}

// ISequenceChunkOnDisk processes a sequence iterator by distributing the sequences
// into chunks stored on disk. It uses a classifier to determine how to distribute
// the sequences and returns a new iterator for the processed sequences.
//
// Parameters:
//   - iterator: An iterator of biosequences to be processed.
//   - classifier: A pointer to a BioSequenceClassifier used to classify the sequences
//     during distribution.
//
// Returns:
// An iterator of biosequences representing the processed chunks. If an error occurs
// during the creation of the temporary directory or any other operation, it returns
// an error along with a nil iterator.
//
// The function operates asynchronously, creating a temporary directory to store
// the sequence chunks. Once the processing is complete, the temporary directory
// is removed. The function logs the number of batches created and the processing
// status of each batch.
func ISequenceChunkOnDisk(iterator obiiter.IBioSequence,
	classifier *obiseq.BioSequenceClassifier) (obiiter.IBioSequence, error) {
	dir, err := tempDir()
	if err != nil {
		return obiiter.NilIBioSequence, err
	}

	newIter := obiiter.MakeIBioSequence()

	newIter.Add(1)

	go func() {
		defer func() {
			os.RemoveAll(dir)
			log.Debugln("Clear the cache directory")
		}()

		newIter.Wait()
		newIter.Close()
	}()

	obiformats.WriterDispatcher(dir+"/chunk_%s.fastx",
		iterator.Distribute(classifier),
		obiformats.WriteSequencesToFile,
	)

	fileNames := find(dir, ".fastx")
	nbatch := len(fileNames)
	log.Infof("Data splitted over %d batches", nbatch)

	go func() {

		for order, file := range fileNames {
			iseq, err := obiformats.ReadSequencesFromFile(file)

			if err != nil {
				panic(err)
			}

			source, chunk := iseq.Load()

			newIter.Push(obiiter.MakeBioSequenceBatch(source, order, chunk))
			log.Infof("Start processing of batch %d/%d : %d sequences",
				order, nbatch, len(chunk))

		}

		newIter.Done()
	}()

	return newIter, err
}
