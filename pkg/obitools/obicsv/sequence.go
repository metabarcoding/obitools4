package obicsv

import (
	"log"
	"slices"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
)

func CSVSequenceHeader(opt Options) CSVHeader {
	keys := opt.CSVKeys()
	record := make(CSVHeader, 0, len(keys)+4)

	if opt.CSVId() {
		record.AppendField("id")
	}

	if opt.CSVCount() {
		record.AppendField("count")
	}

	if opt.CSVTaxon() {
		record.AppendField("taxid")
	}

	if opt.CSVDefinition() {
		record.AppendField("definition")
	}

	for _, field := range opt.CSVKeys() {
		if field != "definition" {
			record.AppendField(field)
		}
	}

	if opt.CSVSequence() {
		record.AppendField("sequence")
	}

	if opt.CSVQuality() {
		record.AppendField("quality")
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

	if opt.CSVAutoColumn() {
		if iter.Next() {
			batch := iter.Get()
			if len(batch.Slice()) == 0 {
				log.Panicf("first batch should not be empty")
			}
			auto_slot := batch.Slice().AttributeKeys(true, true).Members()
			slices.Sort(auto_slot)
			CSVKeys(auto_slot)(opt)
			iter.PushBack()
		}
	}

	newIter := NewICSVRecord()
	newIter.SetHeader(CSVSequenceHeader(opt))

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