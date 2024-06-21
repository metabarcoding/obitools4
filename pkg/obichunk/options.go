package obichunk

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

type __options__ struct {
	statsOn         obiseq.StatsOnDescriptions
	categories      []string
	navalue         string
	cacheOnDisk     bool
	batchCount      int
	batchSize       int
	parallelWorkers int
	noSingleton     bool
}

type Options struct {
	pointer *__options__
}

type WithOption func(Options)

func MakeOptions(setters []WithOption) Options {
	o := __options__{
		statsOn:         make(obiseq.StatsOnDescriptions, 10),
		categories:      make([]string, 0, 100),
		navalue:         "NA",
		cacheOnDisk:     false,
		batchCount:      100,
		batchSize:       obioptions.CLIBatchSize(),
		parallelWorkers: obioptions.CLIParallelWorkers(),
		noSingleton:     false,
	}

	opt := Options{&o}

	for _, set := range setters {
		set(opt)
	}

	return opt
}

func (opt Options) Categories() []string {
	return opt.pointer.categories
}

func (opt Options) PopCategories() string {
	if len(opt.pointer.categories) > 0 {
		c := opt.pointer.categories[0]
		opt.pointer.categories = opt.pointer.categories[1:]
		return c
	}
	return ""
}

func (opt Options) StatsOn() obiseq.StatsOnDescriptions {
	return opt.pointer.statsOn
}

func (opt Options) NAValue() string {
	return opt.pointer.navalue
}

func (opt Options) BatchCount() int {
	return opt.pointer.batchCount
}

func (opt Options) BatchSize() int {
	return opt.pointer.batchSize
}

func (opt Options) ParallelWorkers() int {
	return opt.pointer.parallelWorkers
}

func (opt Options) SortOnDisk() bool {
	return opt.pointer.cacheOnDisk
}

func (opt Options) NoSingleton() bool {
	return opt.pointer.noSingleton
}

func OptionSortOnDisk() WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.cacheOnDisk = true
	})

	return f
}

func OptionSortOnMemory() WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.cacheOnDisk = false
	})

	return f
}
func OptionSubCategory(keys ...string) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.categories = append(opt.pointer.categories, keys...)
	})

	return f
}

func OptionNAValue(na string) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.navalue = na
	})

	return f
}

func OptionStatOn(keys ...string) WithOption {
	f := WithOption(func(opt Options) {
		for _, k := range keys {
			d := obiseq.MakeStatsOnDescription(k)
			opt.pointer.statsOn[d.Name] = d
		}
	})

	return f
}

func OptionBatchCount(number int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.batchCount = number
	})

	return f
}

func OptionsParallelWorkers(nworkers int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.parallelWorkers = nworkers
	})

	return f
}

func OptionsBatchSize(size int) WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.batchSize = size
	})

	return f
}

func OptionsNoSingleton() WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.noSingleton = true
	})

	return f
}

func OptionsWithSingleton() WithOption {
	f := WithOption(func(opt Options) {
		opt.pointer.noSingleton = false
	})

	return f
}
