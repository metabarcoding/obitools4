package obichunk

import (
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiformats"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
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

func ISequenceChunkOnDisk(iterator obiiter.IBioSequenceBatch,
	classifier *obiseq.BioSequenceClassifier,
	sizes ...int) (obiiter.IBioSequenceBatch, error) {
	dir, err := tempDir()
	if err != nil {
		return obiiter.NilIBioSequenceBatch, err
	}

	bufferSize := iterator.BufferSize()

	if len(sizes) > 0 {
		bufferSize = sizes[0]
	}

	newIter := obiiter.MakeIBioSequenceBatch(bufferSize)

	newIter.Add(1)

	go func() {
		defer func() {
			os.RemoveAll(dir)
			log.Println("Clear the cache directory")
		}()

		newIter.Wait()
		newIter.Close()
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

			//chunck := make(obiseq.BioSequenceSlice, 0, 10000)
			chunck := obiseq.MakeBioSequenceSlice()
			for iseq.Next() {
				b := iseq.Get()
				chunck = append(chunck, b.Slice()...)
				b.Recycle()
			}

			newIter.Push(obiiter.MakeBioSequenceBatch(order, chunck))

		}

		newIter.Done()
	}()

	return newIter, err
}
