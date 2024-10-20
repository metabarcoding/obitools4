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

func tempDir() (string, error) {
	dir, err := os.MkdirTemp(os.TempDir(), "obiseq_chunks_")
	if err != nil {
		return "", err
	}
	return dir, nil
}

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
