package obichunk

import (
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

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
	classifier obiseq.SequenceClassifier,
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

	go func() {
		obiformats.WriterDispatcher(dir+"/chunk_%s.fastx",
			iterator.Distribute(classifier),
			obiformats.WriteSequencesBatchToFile,
		)

		files := find(dir, ".fastx")

		for order, file := range files {
			iseq, err := obiformats.ReadSequencesBatchFromFile(file)

			if err != nil {
				panic(err)
			}

			chunck := make(obiseq.BioSequenceSlice, 0, 1000)

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

func ISequenceChunk(iterator obiseq.IBioSequenceBatch,
	classifier obiseq.SequenceClassifier,
	sizes ...int) (obiseq.IBioSequenceBatch, error) {

	bufferSize := iterator.BufferSize()

	if len(sizes) > 0 {
		bufferSize = sizes[0]
	}

	newIter := obiseq.MakeIBioSequenceBatch(bufferSize)

	newIter.Add(1)

	go func() {
		newIter.Wait()
		close(newIter.Channel())
	}()

	go func() {
		lock := sync.Mutex{}

		dispatcher := iterator.Distribute(classifier)

		jobDone := sync.WaitGroup{}
		chunks := make(map[string]*obiseq.BioSequenceSlice, 100)

		for newflux := range dispatcher.News() {
			jobDone.Add(1)
			go func(newflux string) {
				data, err := dispatcher.Outputs(newflux)

				if err != nil {
					log.Fatalf("Cannot retreive the new chanel : %v", err)
				}

				chunk := make(obiseq.BioSequenceSlice, 0, 1000)

				for data.Next() {
					b := data.Get()
					chunk = append(chunk, b.Slice()...)
				}

				lock.Lock()
				chunks[newflux] = &chunk
				lock.Unlock()
				jobDone.Done()
			}(newflux)
		}

		jobDone.Wait()
		order := 0

		for _, chunck := range chunks {

			if len(*chunck) > 0 {
				newIter.Channel() <- obiseq.MakeBioSequenceBatch(order, *chunck...)
				order++
			}

		}
		newIter.Done()
	}()

	return newIter, nil
}
