package obifind

import (
	"slices"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
)

type __options__ struct {
	batch_size           int // Number of items to process in a batch
	with_pattern         bool
	with_parent          bool
	with_path            bool
	with_rank            bool
	with_scientific_name bool
	raw_taxid            bool
	with_metadata        []string
	source               string // Source of the data
}

// Options wraps the __options__ struct to provide a pointer to the options.
type Options struct {
	pointer *__options__ // Pointer to the underlying options
}

// WithOption is a function type that takes an Options parameter and modifies it.
type WithOption func(Options)

// MakeOptions creates an Options instance with default settings and applies any provided setters.
// It returns the configured Options.
//
// Parameters:
//   - setters: A slice of WithOption functions to customize the options.
//
// Returns:
//   - An Options instance with the specified settings.
func MakeOptions(setters []WithOption) Options {
	o := __options__{
		batch_size:           obidefault.BatchSize(), // Number of items to process in a batch
		with_pattern:         true,
		with_parent:          false,
		with_path:            false,
		with_rank:            true,
		with_scientific_name: false,
		raw_taxid:            false,
		source:               "unknown",
	}
	opt := Options{&o}

	for _, set := range setters {
		set(opt)
	}

	return opt
}

// BatchSize returns the size of the batch to be processed.
// It retrieves the batch size from the underlying options.
func (o *Options) BatchSize() int {
	return o.pointer.batch_size
}

// WithPattern returns whether the pattern option is enabled.
// It retrieves the setting from the underlying options.
func (o *Options) WithPattern() bool {
	return o.pointer.with_pattern
}

// WithParent returns whether the parent option is enabled.
// It retrieves the setting from the underlying options.
func (o *Options) WithParent() bool {
	return o.pointer.with_parent
}

// WithPath returns whether the path option is enabled.
// It retrieves the setting from the underlying options.
func (o *Options) WithPath() bool {
	return o.pointer.with_path
}

// WithRank returns whether the rank option is enabled.
// It retrieves the setting from the underlying options.
func (o *Options) WithRank() bool {
	return o.pointer.with_rank
}

// WithScientificName returns whether the scientific name option is enabled.
// It retrieves the setting from the underlying options.
func (o *Options) WithScientificName() bool {
	return o.pointer.with_scientific_name
}

// RawTaxid returns whether the raw taxid option is enabled.
// It retrieves the setting from the underlying options.
func (o *Options) RawTaxid() bool {
	return o.pointer.raw_taxid
}

// Source returns the source of the data.
// It retrieves the source from the underlying options.
func (o *Options) Source() string {
	return o.pointer.source
}

// WithMetadata returns a slice of strings containing the metadata
// associated with the Options instance. It retrieves the metadata
// from the pointer's with_metadata field.
func (o *Options) WithMetadata() []string {
	if o.WithPattern() {
		idx := slices.Index(o.pointer.with_metadata, "query")
		if idx >= 0 {
			o.pointer.with_metadata = slices.Delete(o.pointer.with_metadata, idx, idx+1)
		}
	}

	return o.pointer.with_metadata
}

// OptionsBatchSize returns a WithOption function that sets the batch_size option.
// Parameters:
//   - size: An integer specifying the size of the batch to be processed.
func OptionsBatchSize(size int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.batch_size = size
	})

	return f
}

func OptionsWithPattern(value bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.with_pattern = value
	})

	return f
}

func OptionsWithParent(value bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.with_parent = value
	})

	return f
}

func OptionsWithPath(value bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.with_path = value
	})

	return f
}

func OptionsWithRank(value bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.with_rank = value
	})

	return f
}

func OptionsWithScientificName(value bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.with_scientific_name = value
	})

	return f
}

func OptionsRawTaxid(value bool) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.raw_taxid = value
	})

	return f
}

func OptionsSource(value string) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.source = value
	})

	return f
}

func OptionsWithMetadata(values ...string) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.with_metadata = values
	})
	return f
}

func NewCSVTaxaIterator(iterator *obitax.ITaxon, options ...WithOption) *obiiter.ICSVRecord {

	opt := MakeOptions(options)
	metakeys := make([]string, 0)

	newIter := obiiter.NewICSVRecord()

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
		data := make([]obiiter.CSVRecord, 0, batch_size)
		for iterator.Next() {

			taxon := iterator.Get()
			record := make(obiiter.CSVRecord)

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
				newIter.Push(obiiter.MakeCSVRecordBatch(opt.Source(), o, data))
				data = make([]obiiter.CSVRecord, 0, batch_size)
				o++
			}

		}

		if len(data) > 0 {
			newIter.Push(obiiter.MakeCSVRecordBatch(opt.Source(), o, data))
		}

		newIter.Done()
	}()

	return newIter
}
