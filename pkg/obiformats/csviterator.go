package obiformats

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiitercsv"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
)

func CSVTaxaIterator(iterator *obitax.ITaxon, options ...WithOption) *obiitercsv.ICSVRecord {

	opt := MakeOptions(options)
	metakeys := make([]string, 0)

	newIter := obiitercsv.NewICSVRecord()

	newIter.Add(1)

	batch_size := opt.BatchSize()

	if opt.WithPattern() {
		newIter.AppendField("query")
		opt.pointer.with_metadata = append(opt.pointer.with_metadata, "query")
	}

	newIter.AppendField("taxid")
	rawtaxid := opt.RawTaxid()

	if opt.WithParent() {
		newIter.AppendField("parent")
	}

	if opt.WithRank() {
		newIter.AppendField("taxonomic_rank")
	}

	if opt.WithScientificName() {
		newIter.AppendField("scientific_name")
	}

	if opt.WithMetadata() != nil {
		metakeys = opt.WithMetadata()
		for _, metadata := range metakeys {
			newIter.AppendField(metadata)
		}
	}

	if opt.WithPath() {
		newIter.AppendField("path")
	}

	go func() {
		newIter.WaitAndClose()
	}()

	go func() {
		o := 0
		data := make([]obiitercsv.CSVRecord, 0, batch_size)
		for iterator.Next() {

			taxon := iterator.Get()
			record := make(obiitercsv.CSVRecord)

			if opt.WithPattern() {
				record["query"] = taxon.MetadataAsString("query")
			}

			if rawtaxid {
				record["taxid"] = *taxon.Node.Id()
			} else {
				record["taxid"] = taxon.String()
			}

			if opt.WithParent() {
				if rawtaxid {
					record["parent"] = *taxon.Node.ParentId()
				} else {
					record["parent"] = taxon.Parent().String()
				}
			}

			if opt.WithRank() {
				record["taxonomic_rank"] = taxon.Rank()
			}

			if opt.WithScientificName() {
				record["scientific_name"] = taxon.ScientificName()
			}

			if opt.WithPath() {
				record["path"] = taxon.Path().String()
			}

			for _, key := range metakeys {
				record[key] = taxon.MetadataAsString(key)
			}

			data = append(data, record)
			if len(data) >= batch_size {
				newIter.Push(obiitercsv.MakeCSVRecordBatch(opt.Source(), o, data))
				data = make([]obiitercsv.CSVRecord, 0, batch_size)
				o++
			}

		}

		if len(data) > 0 {
			newIter.Push(obiitercsv.MakeCSVRecordBatch(opt.Source(), o, data))
		}

		newIter.Done()
	}()

	return newIter
}
