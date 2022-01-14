package obiformats

// #cgo CFLAGS: -g -Wall
// #cgo LDFLAGS: -lz
// #include <stdlib.h>
// #include "fastseq_read.h"
import "C"

import (
	"fmt"
	"log"
	"os"
	"time"
	"unsafe"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/cutils"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func __fastseq_reader__(seqfile C.fast_kseq_p,
	iterator obiseq.IBioSequenceBatch,
	batch_size int) {
	var comment string
	i := 0
	ii := 0

	slice := make(obiseq.BioSequenceSlice, 0, batch_size)

	for l := int64(C.next_fast_sek(seqfile)); l > 0; l = int64(C.next_fast_sek(seqfile)) {

		s := seqfile.seq

		sequence := C.GoBytes(unsafe.Pointer(s.seq.s),
			C.int(s.seq.l))

		name := C.GoString(s.name.s)

		if s.comment.l > C.ulong(0) {
			comment = C.GoString(s.comment.s)
		} else {
			comment = ""
		}

		rep := obiseq.MakeBioSequence(name, sequence, comment)

		if s.qual.l > C.ulong(0) {
			cquality := cutils.ByteSlice(unsafe.Pointer(s.qual.s), int(s.qual.l))
			quality := make(obiseq.Quality, s.qual.l)
			l := int(s.qual.l)
			shift := uint8(seqfile.shift)
			for j := 0; j < l; j++ {
				quality[j] = uint8(cquality[j]) - shift
			}

			rep.SetQualities(quality)
		}
		slice = append(slice, rep)
		ii++
		if ii >= batch_size {
			// log.Printf("\n==> Pushing sequence batch\n")
			// start := time.Now()

			iterator.Channel() <- obiseq.MakeBioSequenceBatch(i, slice...)
			// elapsed := time.Since(start)
			// log.Printf("\n==>sequences pushed after %s\n", elapsed)

			slice = make(obiseq.BioSequenceSlice, 0, batch_size)
			i++
			ii = 0
		}
	}
	if len(slice) > 0 {
		iterator.Channel() <- obiseq.MakeBioSequenceBatch(i, slice...)
	}
	iterator.Done()

}

func ReadFastSeqBatchFromFile(filename string, options ...WithOption) (obiseq.IBioSequenceBatch, error) {
	opt := MakeOptions(options)

	name := C.CString(filename)
	defer C.free(unsafe.Pointer(name))

	pointer := C.open_fast_sek_file(name, C.int32_t(opt.QualityShift()))

	var err error
	err = nil

	if pointer == nil {
		err = fmt.Errorf("cannot open file %s", filename)
		return obiseq.NilIBioSequenceBatch, err
	}

	size := int64(-1)
	fi, err := os.Stat(filename)
	if err == nil {
		size = fi.Size()
		log.Printf("File size of %s is %d bytes\n", filename, size)
	} else {
		size = -1
	}

	new_iter := obiseq.MakeIBioSequenceBatch(opt.BufferSize())
	new_iter.Add(1)

	go func() {
		new_iter.Wait()
		for len(new_iter.Channel()) > 0 {
			time.Sleep(time.Millisecond)
		}
		close(new_iter.Channel())

		log.Println("End of the fastq file reading")
	}()

	log.Println("Start of the fastq file reading")

	go __fastseq_reader__(pointer, new_iter, opt.BatchSize())
	parser := opt.ParseFastSeqHeader()
	if parser != nil {
		return IParseFastSeqHeaderBatch(new_iter, options...), err
	}

	return new_iter, err
}

func ReadFastSeqFromFile(filename string, options ...WithOption) (obiseq.IBioSequence, error) {
	ib, err := ReadFastSeqBatchFromFile(filename, options...)
	return ib.SortBatches().IBioSequence(), err
}

func ReadFastSeqBatchFromStdin(options ...WithOption) obiseq.IBioSequenceBatch {
	opt := MakeOptions(options)
	new_iter := obiseq.MakeIBioSequenceBatch(opt.BufferSize())

	new_iter.Add(1)

	go func() {
		new_iter.Wait()
		close(new_iter.Channel())
	}()

	go __fastseq_reader__(C.open_fast_sek_stdin(C.int32_t(opt.QualityShift())), new_iter, opt.BatchSize())

	return new_iter
}

func ReadFastSeqFromStdin(options ...WithOption) obiseq.IBioSequence {
	ib := ReadFastSeqBatchFromStdin(options...)
	return ib.SortBatches().IBioSequence()
}
