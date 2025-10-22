package obiutils

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"regexp"

	"github.com/gabriel-vasile/mimetype"
	log "github.com/sirupsen/logrus"
)

func HasBOM(data []byte) bool {
	switch {
	case bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}):
		log.Infoln("BOM detected: UTF-8 (EF BB BF)")
		return true
	case bytes.HasPrefix(data, []byte{0xFE, 0xFF}):
		log.Infoln("BOM detected: UTF-16 Big Endian (FE FF)")
		return true
	case bytes.HasPrefix(data, []byte{0xFF, 0xFE}):
		log.Infoln("BOM detected: UTF-16 Little Endian (FF FE)")
		return true
	case bytes.HasPrefix(data, []byte{0x00, 0x00, 0xFE, 0xFF}):
		log.Infoln("BOM detected: UTF-32 Big Endian (00 00 FE FF)")
		return true
	case bytes.HasPrefix(data, []byte{0xFF, 0xFE, 0x00, 0x00}):
		log.Infoln("BOM detected: UTF-32 Little Endian (FF FE 00 00)")
		return true
	default:
		log.Infoln("No BOM detected")
		return false
	}
}

func DropLastLine(b []byte) []byte {
	for i := len(b) - 1; i > 0; i-- {
		if b[i] == '\n' {
			return b[:i]
		}
	}
	return b
}

var __obimimetype_registred__ = false

func RegisterOBIMimeType() {
	if !__obimimetype_registred__ {
		csv := func(in []byte, limit uint32) bool {
			in = DropLastLine(in)

			br := bytes.NewReader(in)
			r := csv.NewReader(br)
			r.Comma = ','
			r.ReuseRecord = true
			r.LazyQuotes = true
			r.Comment = '#'

			lines := 0
			for {
				_, err := r.Read()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					return false
				}
				lines++
			}

			return r.FieldsPerRecord > 1 && lines > 1
		}

		fastaDetector := func(raw []byte, limit uint32) bool {
			ok, err := regexp.Match("^>[^ ]", raw)
			return ok && err == nil
		}

		fastqDetector := func(raw []byte, limit uint32) bool {
			ok, err := regexp.Match("^@[^ ].*\n[A-Za-z.-]+", raw)
			if ok && err == nil {
				ok, err = regexp.Match("^@param,[^ ]", raw)
				ok = !ok
			}
			return ok && err == nil
		}

		ecoPCR2Detector := func(raw []byte, limit uint32) bool {
			ok := bytes.HasPrefix(raw, []byte("#@ecopcr-v2"))
			return ok
		}

		genbankDetector := func(raw []byte, limit uint32) bool {
			ok2 := bytes.HasPrefix(raw, []byte("LOCUS       "))
			ok1, err := regexp.Match("^[^ ]* +Genetic Sequence Data Bank *\n", raw)
			return ok2 || (ok1 && err == nil)
		}

		emblDetector := func(raw []byte, limit uint32) bool {
			ok := bytes.HasPrefix(raw, []byte("ID   "))
			return ok
		}

		mimetype.Lookup("text/plain").Extend(fastaDetector, "text/fasta", ".fasta")
		mimetype.Lookup("text/plain").Extend(fastqDetector, "text/fastq", ".fastq")
		mimetype.Lookup("text/plain").Extend(ecoPCR2Detector, "text/ecopcr2", ".ecopcr")
		mimetype.Lookup("text/plain").Extend(genbankDetector, "text/genbank", ".seq")
		mimetype.Lookup("text/plain").Extend(emblDetector, "text/embl", ".dat")
		mimetype.Lookup("text/plain").Extend(csv, "text/csv", ".csv")

		mimetype.Lookup("application/octet-stream").Extend(fastaDetector, "text/fasta", ".fasta")
		mimetype.Lookup("application/octet-stream").Extend(fastqDetector, "text/fastq", ".fastq")
		mimetype.Lookup("application/octet-stream").Extend(ecoPCR2Detector, "text/ecopcr2", ".ecopcr")
		mimetype.Lookup("application/octet-stream").Extend(genbankDetector, "text/genbank", ".seq")
		mimetype.Lookup("application/octet-stream").Extend(emblDetector, "text/embl", ".dat")
		mimetype.Lookup("application/octet-stream").Extend(csv, "text/csv", ".csv")
	}
	__obimimetype_registred__ = true
}
