package obicsv

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"

	log "github.com/sirupsen/logrus"
)

func CSVSequenceHeader(opt Options) CSVHeader {
	keys := opt.CSVKeys()
	record := make([]string, 0, len(keys)+4)

	if opt.CSVId() {
		record = append(record, "id")
	}

	if opt.CSVCount() {
		record = append(record, "count")
	}

	if opt.CSVTaxon() {
		record = append(record, "taxid")
	}

	if opt.CSVDefinition() {
		record = append(record, "definition")
	}

	record = append(record, opt.CSVKeys()...)

	if opt.CSVSequence() {
		record = append(record, "sequence")
	}

	if opt.CSVQuality() {
		record = append(record, "quality")
	}

	return record
}

func CSVBatchFromSequences(batch obiiter.BioSequenceBatch, opt Options) CSVRecordBatch {
	keys := opt.CSVKeys()
	csvslice := make([]CSVRecord, batch.Len())

	for i, sequence := range batch.Slice() {
		record := make(CSVRecord)

		if opt.CSVId() {
			record["id"] = sequence.Id()
		}

		if opt.CSVCount() {
			record["count"] = sequence.Count()
		}

		if opt.CSVTaxon() {
			var taxid string
			taxon := sequence.Taxon(nil)

			if taxon != nil {
				taxid = taxon.String()
			} else {
				taxid = sequence.Taxid()
			}

			record["taxid"] = taxid
		}

		if opt.CSVDefinition() {
			record["definition"] = sequence.Definition()
		}

		for _, key := range keys {
			value, ok := sequence.GetAttribute(key)
			if !ok {
				value = opt.CSVNAValue()
			}

			record[key] = value
		}

		if opt.CSVSequence() {
			record["sequence"] = string(sequence.Sequence())
		}

		if opt.CSVQuality() {
			if sequence.HasQualities() {
				l := sequence.Len()
				q := sequence.Qualities()
				ascii := make([]byte, l)
				quality_shift := obioptions.OutputQualityShift()
				for j := 0; j < l; j++ {
					ascii[j] = uint8(q[j]) + uint8(quality_shift)
				}
				record["quality"] = string(ascii)
			} else {
				record["quality"] = opt.CSVNAValue()
			}
		}

		csvslice[i] = record
	}

	return MakeCSVRecordBatch(batch.Source(), batch.Order(), csvslice)
}

func NewCSVSequenceIterator(iter obiiter.IBioSequence, options ...WithOption) *ICSVRecord {

	opt := MakeOptions(options)

	newIter := NewICSVRecord()
	newIter.SetHeader(CSVSequenceHeader(opt))

	log.Warnf("", newIter.Header())

	nwriters := opt.ParallelWorkers()
	newIter.Add(nwriters)

	go func() {
		newIter.WaitAndClose()
	}()

	ff := func(iterator obiiter.IBioSequence) {
		for iterator.Next() {

			batch := iterator.Get()
			newIter.Push(CSVBatchFromSequences(batch, opt))
		}
		newIter.Done()
	}

	go ff(iter)
	for i := 0; i < nwriters-1; i++ {
		go ff(iter.Split())
	}

	return newIter
}
