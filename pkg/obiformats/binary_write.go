package obiformats

import (
	"io"
	"log"
	"os"
	"time"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"github.com/renproject/surge"
)

func WriteBinary(iterator obiseq.IBioSequence, file io.Writer) error {
	singleseq := make(obiseq.BioSequenceSlice, 1)
	blob := make([]byte, 0, 1024)
	for iterator.Next() {
		singleseq[0] = iterator.Get()
		blobsize := singleseq.SizeHint()
		if blobsize > cap(blob) {
			blob = make([]byte, 0, blobsize*2+8)
		}
		_, _, err := surge.MarshalI64(int64(blobsize), blob, 8)
		if err != nil {
			return err
		}
		data := blob[8 : 8+blobsize]
		_, _, err = singleseq.Marshal(data, blobsize)
		if err != nil {
			return err
		}

		file.Write(blob[0 : 8+blobsize])
	}

	return nil
}

func WriteBinaryToFile(iterator obiseq.IBioSequence,
	filename string) error {

	file, err := os.Create(filename)

	if err != nil {
		log.Fatalf("open file error: %v", err)
		return err
	}

	return WriteBinary(iterator, file)
}

func WriteBinaryToStdout(iterator obiseq.IBioSequence) error {
	return WriteBinary(iterator, os.Stdout)
}

func WriteBinaryBatch(iterator obiseq.IBioSequenceBatch, file io.Writer, options ...WithOption) (obiseq.IBioSequenceBatch, error) {
	opt := MakeOptions(options)

	buffsize := iterator.BufferSize()
	newIter := obiseq.MakeIBioSequenceBatch(buffsize)

	nwriters := opt.ParallelWorkers()

	chunkchan := make(chan FileChunck)

	newIter.Add(nwriters)

	go func() {
		newIter.Wait()
		for len(chunkchan) > 0 {
			time.Sleep(time.Millisecond)
		}
		close(chunkchan)
		for len(newIter.Channel()) > 0 {
			time.Sleep(time.Millisecond)
		}
		close(newIter.Channel())
	}()

	ff := func(iterator obiseq.IBioSequenceBatch) {
		blob := make([]byte, 0, 10240)
		for iterator.Next() {
			batch := iterator.Get()
			blobsize := batch.Slice().SizeHint()
			if blobsize > cap(blob) {
				blob = make([]byte, 0, blobsize*2+8)
			}
			_, _, err := surge.MarshalI64(int64(blobsize), blob, 8)
			if err != nil {
				log.Fatalf("error in reading binary file %v\n", err)
			}
			data := blob[8 : 8+blobsize]
			_, _, err = batch.Slice().Marshal(data, blobsize)
			if err != nil {
				log.Fatalf("error in reading binary file %v\n", err)
			}

			chunkchan <- FileChunck{
				data,
				batch.Order(),
			}
			newIter.Channel() <- batch
		}
		newIter.Done()
	}

	log.Println("Start of the binary file writing")
	go ff(iterator)
	for i := 0; i < nwriters-1; i++ {
		go ff(iterator.Split())
	}

	next_to_send := 0
	received := make(map[int]FileChunck, 100)

	go func() {
		for chunk := range chunkchan {
			if chunk.order == next_to_send {
				file.Write(chunk.text)
				next_to_send++
				chunk, ok := received[next_to_send]
				for ok {
					file.Write(chunk.text)
					delete(received, next_to_send)
					next_to_send++
					chunk, ok = received[next_to_send]
				}
			} else {
				received[chunk.order] = chunk
			}

		}
	}()

	return newIter, nil
}

func WriteBinaryBatchToStdout(iterator obiseq.IBioSequenceBatch, options ...WithOption) (obiseq.IBioSequenceBatch, error) {
	return WriteBinaryBatch(iterator, os.Stdout, options...)
}

func WriteBinaryBatchToFile(iterator obiseq.IBioSequenceBatch,
	filename string,
	options ...WithOption) (obiseq.IBioSequenceBatch, error) {

	file, err := os.Create(filename)

	if err != nil {
		log.Fatalf("open file error: %v", err)
		return obiseq.NilIBioSequenceBatch, err
	}

	return WriteBinaryBatch(iterator, file, options...)
}
