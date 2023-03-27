package obiformats

// #cgo CFLAGS: -g -Wall
// #cgo LDFLAGS: -lz
// #include <stdlib.h>
// #include "fastseq_read.h"
import "C"

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"unsafe"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiutils"
)

func _FastseqReader(source string,
	seqfile C.fast_kseq_p,
	iterator obiiter.IBioSequence,
	batch_size int) {
	var comment string
	i := 0
	ii := 0

	slice := obiseq.MakeBioSequenceSlice()

	for l := int64(C.next_fast_sek(seqfile)); l != 0; l = int64(C.next_fast_sek(seqfile)) {
		s := seqfile.seq

		sequence := C.GoBytes(unsafe.Pointer(s.seq.s), C.int(s.seq.l))
		name := C.GoString(s.name.s)

		if s.comment.l > C.ulong(0) {
			comment = C.GoString(s.comment.s)
		} else {
			comment = ""
		}

		rep := obiseq.NewBioSequence(name, bytes.ToLower(sequence), comment)
		rep.SetSource(source)
		if s.qual.l > C.ulong(0) {
			cquality := unsafe.Slice(s.qual.s, C.int(s.qual.l))
			l := int(s.qual.l)
			quality := obiseq.GetSlice(l)
			shift := uint8(seqfile.shift)

			for j := 0; j < l; j++ {
				func() {
					defer func() {
						if err := recover(); err != nil {
							log.Println("cquality:", cquality,
								"s.qual.s:", s.qual.s,
								"quality:", quality)
							log.Panic("panic occurred:", err)
						}
					}()
					quality = append(quality, uint8(cquality[j])-shift)
				}()
			}

			rep.SetQualities(quality)
		}
		slice = append(slice, rep)
		ii++
		if ii >= batch_size {
			iterator.Push(obiiter.MakeBioSequenceBatch(i, slice))
			slice = obiseq.MakeBioSequenceSlice()
			i++
			ii = 0
		}

	}
	if len(slice) > 0 {
		iterator.Push(obiiter.MakeBioSequenceBatch(i, slice))
	}
	iterator.Done()

}

func ReadFastSeqFromFile(filename string, options ...WithOption) (obiiter.IBioSequence, error) {

	options = append(options, OptionsSource(obiutils.RemoveAllExt((path.Base(filename)))))

	opt := MakeOptions(options)

	name := C.CString(filename)
	defer C.free(unsafe.Pointer(name))

	pointer := C.open_fast_sek_file(name, C.int32_t(opt.QualityShift()))

	var err error
	err = nil

	if pointer == nil {
		err = fmt.Errorf("cannot open file %s", filename)
		return obiiter.NilIBioSequence, err
	}

	size := int64(-1)
	fi, err := os.Stat(filename)
	if err == nil {
		size = fi.Size()
		log.Debugf("File size of %s is %d bytes\n", filename, size)
	} else {
		size = -1
	}

	newIter := obiiter.MakeIBioSequence()
	newIter.Add(1)

	go func(iter obiiter.IBioSequence) {
		iter.WaitAndClose()
		log.Debugln("End of the fastx file reading")
	}(newIter)

	log.Debugln("Start of the fastx file reading")

	go _FastseqReader(opt.Source(), pointer, newIter, opt.BatchSize())

	log.Debugln("Full file batch mode : ", opt.FullFileBatch())
	if opt.FullFileBatch() {
		newIter = newIter.FullFileIterator()
	}

	parser := opt.ParseFastSeqHeader()

	if parser != nil {
		return IParseFastSeqHeaderBatch(newIter, options...), err
	}

	return newIter, err
}

func ReadFastSeqFromStdin(options ...WithOption) obiiter.IBioSequence {

	options = append(options, OptionsSource("stdin"))

	opt := MakeOptions(options)
	newIter := obiiter.MakeIBioSequence()

	newIter.Add(1)

	go func(iter obiiter.IBioSequence) {
		iter.WaitAndClose()
	}(newIter)

	go _FastseqReader(opt.Source(),
		C.open_fast_sek_stdin(C.int32_t(opt.QualityShift())),
		newIter, opt.BatchSize())

	log.Debugln("Full file batch mode : ", opt.FullFileBatch())
	if opt.FullFileBatch() {
		newIter = newIter.FullFileIterator()
	}

	parser := opt.ParseFastSeqHeader()

	if parser != nil {
		return IParseFastSeqHeaderBatch(newIter, options...)
	}

	return newIter
}
