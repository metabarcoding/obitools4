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

func ISequenceChunk(iterator obiseq.IBioSequenceBatch, size int, sizes ...int) (obiseq.IBioSequenceBatch, error) {
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
		newIter.Wait()
		close(newIter.Channel())
		log.Println("====>> clear diectory")
		os.RemoveAll(dir)
	}()

	go func() {
		obiformats.WriterDispatcher(dir+"/chunk_%s.fastx",
			iterator.Distribute(obiseq.HashClassifier(size)),
			obiformats.WriteSequencesBatchToFile,
		)

		files := find(dir, ".fastx")

		for order, file := range files {
			iseq, err := obiformats.ReadSequencesBatchFromFile(file)

			if err != nil {
				panic(err)
			}

			chunck := make(obiseq.BioSequenceSlice, 0, 3*size)

			for iseq.Next() {
				b := iseq.Get()
				chunck = append(chunck, b.Slice()...)
			}

			if len(chunck) > 0 {
				newIter.Channel() <- obiseq.MakeBioSequenceBatch(order, chunck...)
			}

		}

		newIter.Done()
	}()

	return newIter, err
}
