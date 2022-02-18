package obichunk

import (
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiformats"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func tempDir() (string, error) {
	dir, err := ioutil.TempDir(os.TempDir(), "obiseq_chunks_")
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

func ISequenceChunkOnDisk(iterator obiseq.IBioSequenceBatch,
	classifier *obiseq.BioSequenceClassifier,
	sizes ...int) (obiseq.IBioSequenceBatch, error) {
	dir, err := tempDir()
	if err != nil {
		return obiseq.NilIBioSequenceBatch, err
	}

	bufferSize := iterator.BufferSize()

	if len(sizes) > 0 {
		bufferSize = sizes[0]
	}

	newIter := obiseq.MakeIBioSequenceBatch(bufferSize)

	newIter.Add(1)

	go func() {
		defer func() {
			os.RemoveAll(dir)
			log.Println("Clear the cache directory")
		}()

		newIter.Wait()
		close(newIter.Channel())
	}()

	obiformats.WriterDispatcher(dir+"/chunk_%s.fastx",
		iterator.Distribute(classifier),
		obiformats.WriteSequencesBatchToFile,
	)

	fileNames := find(dir, ".fastx")
	log.Println("batch count ", len(fileNames))

	go func() {

		for order, file := range fileNames {
			iseq, err := obiformats.ReadSequencesBatchFromFile(file)

			if err != nil {
				panic(err)
			}

			chunck := make(obiseq.BioSequenceSlice, 0, 10000)

			for iseq.Next() {
				b := iseq.Get()
				chunck = append(chunck, b.Slice()...)
			}

			newIter.Channel() <- obiseq.MakeBioSequenceBatch(order, chunck...)

		}

		newIter.Done()
	}()

	return newIter, err
}
