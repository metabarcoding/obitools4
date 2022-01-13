package obiformats

import (
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

type __ecopcr_file__ struct {
	file           io.Reader
	csv            *csv.Reader
	names          map[string]int
	version        int
	mode           string
	forward_primer string
	reverse_primer string
}

func __readline__(stream io.Reader) string {
	line := make([]byte, 1024)
	char := make([]byte, 1)

	i := 0
	for n, err := stream.Read(char); err == nil && n == 1 && char[0] != '\n'; n, err = stream.Read(char) {
		line[i] = char[0]
		i++
	}
	return string(line[0:i])
}

func __read_ecopcr_bioseq__(file *__ecopcr_file__) (obiseq.BioSequence, error) {

	record, err := file.csv.Read()

	if err != nil {
		return obiseq.NilBioSequence, err
	}

	name := strings.TrimSpace(record[0])

	// Ensure that sequence name is unique accross a file.
	if val, ok := file.names[name]; ok {
		file.names[name]++
		name = fmt.Sprintf("%s_%d", name, val)
	} else {
		file.names[name] = 1
	}

	var sequence []byte
	var comment string

	if file.version == 2 {
		sequence = []byte(strings.TrimSpace(record[20]))
		comment = strings.TrimSpace(record[21])

	} else {
		sequence = []byte(strings.TrimSpace(record[18]))
		comment = strings.TrimSpace(record[19])
	}

	bseq := obiseq.MakeBioSequence(name, sequence, comment)
	annotation := bseq.Annotations()

	annotation["ac"] = name
	annotation["seq_length"], _ = strconv.Atoi(strings.TrimSpace(record[1]))
	annotation["taxid"], _ = strconv.Atoi(strings.TrimSpace(record[2]))
	annotation["rank"] = strings.TrimSpace(record[3])
	annotation["species_taxid"], _ = strconv.Atoi(strings.TrimSpace(record[4]))
	annotation["species_name"] = strings.TrimSpace(record[5])
	annotation["genus_taxid"], _ = strconv.Atoi(strings.TrimSpace(record[6]))
	annotation["genus_name"] = strings.TrimSpace(record[7])
	annotation["family_taxid"], _ = strconv.Atoi(strings.TrimSpace(record[8]))
	annotation["family_name"] = strings.TrimSpace(record[9])
	k_m_taxid := file.mode + "_taxid"
	k_m_name := file.mode + "_name"
	annotation[k_m_taxid], _ = strconv.Atoi(strings.TrimSpace(record[10]))
	annotation[k_m_name] = strings.TrimSpace(record[11])
	annotation["strand"] = strings.TrimSpace(record[12])
	annotation["forward_primer"] = file.forward_primer
	annotation["forward_match"] = strings.TrimSpace(record[13])
	annotation["forward_mismatch"], _ = strconv.Atoi(strings.TrimSpace(record[14]))

	delta := 0
	if file.version == 2 {
		value, err := strconv.ParseFloat(strings.TrimSpace(record[15]), 64)
		if err != nil {
			annotation["forward_tm"] = value
		} else {
			annotation["forward_tm"] = -1
		}
		delta++
	}

	annotation["reverse_primer"] = file.reverse_primer
	annotation["reverse_match"] = strings.TrimSpace(record[15+delta])
	annotation["reverse_mismatch"], _ = strconv.Atoi(strings.TrimSpace(record[16+delta]))

	if file.version == 2 {
		value, err := strconv.ParseFloat(strings.TrimSpace(record[17+delta]), 64)
		if err != nil {
			annotation["reverse_tm"] = value
		} else {
			annotation["reverse_tm"] = -1
		}
		delta++
	}

	annotation["amplicon_length"], _ = strconv.Atoi(strings.TrimSpace(record[17+delta]))

	return bseq, nil
}

func ReadEcoPCRBatch(reader io.Reader, options ...WithOption) obiseq.IBioSequenceBatch {
	tag := make([]byte, 11)
	n, _ := reader.Read(tag)

	version := 1
	if n == 11 && string(tag) == "#@ecopcr-v2" {
		version = 2
	}

	line := __readline__(reader)
	for !strings.HasPrefix(line, "# direct  strand oligo1") {
		line = __readline__(reader)
	}
	forward_primer := (strings.Split(line, " "))[6]

	line = __readline__(reader)
	for !strings.HasPrefix(line, "# reverse strand oligo2") {
		line = __readline__(reader)
	}
	reverse_primer := (strings.Split(line, " "))[5]

	line = __readline__(reader)
	for !strings.HasPrefix(line, "# output in") {
		line = __readline__(reader)
	}
	mode := (strings.Split(line, " "))[3]

	file := csv.NewReader(reader)
	file.Comma = '|'
	file.Comment = '#'
	file.TrimLeadingSpace = true
	file.ReuseRecord = true

	log.Printf("EcoPCR file version : %d  Mode : %s\n", version, mode)

	ecopcr := __ecopcr_file__{
		file:           reader,
		csv:            file,
		names:          make(map[string]int),
		version:        version,
		mode:           mode,
		forward_primer: forward_primer,
		reverse_primer: reverse_primer}

	opt := MakeOptions(options)

	new_iter := obiseq.MakeIBioSequenceBatch(opt.BufferSize())
	new_iter.Add(1)

	go func() {
		new_iter.Wait()
		close(new_iter.Channel())
	}()

	go func() {

		seq, err := __read_ecopcr_bioseq__(&ecopcr)
		slice := make(obiseq.BioSequenceSlice, 0, opt.BatchSize())
		i := 0
		ii := 0
		for err == nil {
			slice = append(slice, seq)
			ii++
			if ii >= opt.BatchSize() {
				new_iter.Channel() <- obiseq.MakeBioSequenceBatch(i, slice...)
				slice = make(obiseq.BioSequenceSlice, 0, opt.BatchSize())

				i++
				ii = 0
			}

			seq, err = __read_ecopcr_bioseq__(&ecopcr)
		}

		if len(slice) > 0 {
			new_iter.Channel() <- obiseq.MakeBioSequenceBatch(i, slice...)
		}

		new_iter.Done()

		if err != nil && err != io.EOF {
			log.Panicf("%+v", err)
		}

	}()

	return new_iter
}

func ReadEcoPCR(reader io.Reader, options ...WithOption) obiseq.IBioSequence {
	ib := ReadEcoPCRBatch(reader, options...)
	return ib.SortBatches().IBioSequence()
}

func ReadEcoPCRBatchFromFile(filename string, options ...WithOption) (obiseq.IBioSequenceBatch, error) {
	var reader io.Reader
	var greader io.Reader
	var err error

	reader, err = os.Open(filename)
	if err != nil {
		log.Printf("open file error: %+v", err)
		return obiseq.NilIBioSequenceBatch, err
	}

	// Test if the flux is compressed by gzip
	greader, err = gzip.NewReader(reader)
	if err == nil {
		reader = greader
	}

	return ReadEcoPCRBatch(reader, options...), nil
}

func ReadEcoPCRFromFile(filename string, options ...WithOption) (obiseq.IBioSequence, error) {
	ib, err := ReadEcoPCRBatchFromFile(filename, options...)
	return ib.SortBatches().IBioSequence(), err

}
