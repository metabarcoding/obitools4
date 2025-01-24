package obisummary

import (
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

type DataSummary struct {
	read_count          int
	variant_count       int
	symbole_count       int
	has_merged_sample   int
	has_obiclean_status int
	has_obiclean_weight int
	tags                map[string]int
	map_tags            map[string]int
	vector_tags         map[string]int
	samples             map[string]int
	sample_variants     map[string]int
	sample_singletons   map[string]int
	sample_obiclean_bad map[string]int
	map_summaries       map[string]map[string]int
}

func NewDataSummary() *DataSummary {
	return &DataSummary{
		read_count:          0,
		variant_count:       0,
		symbole_count:       0,
		has_merged_sample:   0,
		has_obiclean_status: 0,
		has_obiclean_weight: 0,
		tags:                make(map[string]int),
		map_tags:            make(map[string]int),
		vector_tags:         make(map[string]int),
		samples:             make(map[string]int),
		sample_variants:     make(map[string]int),
		sample_singletons:   make(map[string]int),
		sample_obiclean_bad: make(map[string]int),
		map_summaries:       make(map[string]map[string]int),
	}
}

func sumUpdateIntMap(m1, m2 map[string]int) map[string]int {
	for k, v2 := range m2 {
		if v1, ok := m1[k]; ok {
			m1[k] = v1 + v2
		} else {
			m1[k] = v2
		}
	}

	return m1
}

func countUpdateIntMap(m1, m2 map[string]int) map[string]int {
	for k := range m2 {
		if v, ok := m1[k]; ok {
			m1[k] = v + 1
		} else {
			m1[k] = 1
		}
	}

	return m1
}

func plusUpdateIntMap(m1 map[string]int, key string, val int) map[string]int {
	if v, ok := m1[key]; ok {
		m1[key] = v + val
	} else {
		m1[key] = val
	}
	return m1
}

func plusOneUpdateIntMap(m1 map[string]int, key string) map[string]int {
	return plusUpdateIntMap(m1, key, 1)
}

func (data1 *DataSummary) Add(data2 *DataSummary) *DataSummary {
	rep := NewDataSummary()
	rep.read_count = data1.read_count + data2.read_count
	rep.variant_count = data1.variant_count + data2.variant_count
	rep.symbole_count = data1.symbole_count + data2.symbole_count
	rep.has_merged_sample = data1.has_merged_sample + data2.has_merged_sample
	rep.has_obiclean_status = data1.has_obiclean_status + data2.has_obiclean_status
	rep.has_obiclean_weight = data1.has_obiclean_weight + data2.has_obiclean_weight

	rep.tags = sumUpdateIntMap(data1.tags, data2.tags)
	rep.map_tags = sumUpdateIntMap(data1.map_tags, data2.map_tags)
	rep.vector_tags = sumUpdateIntMap(data1.vector_tags, data2.vector_tags)
	rep.samples = sumUpdateIntMap(data1.samples, data2.samples)
	rep.sample_variants = sumUpdateIntMap(data1.sample_variants, data2.sample_variants)
	rep.sample_singletons = sumUpdateIntMap(data1.sample_singletons, data2.sample_singletons)
	rep.sample_obiclean_bad = sumUpdateIntMap(data1.sample_obiclean_bad, data2.sample_obiclean_bad)

	return rep
}

func (data *DataSummary) Update(s *obiseq.BioSequence) *DataSummary {
	data.read_count += s.Count()
	data.variant_count++
	data.symbole_count += s.Len()

	if s.HasAttribute("merged_sample") {
		data.has_merged_sample++
		samples, _ := s.GetIntMap("merged_sample")
		obiclean, obc_ok := s.GetStringMap("obiclean_status")
		data.samples = sumUpdateIntMap(data.samples, samples)
		data.sample_variants = countUpdateIntMap(data.sample_variants, samples)
		for k, v := range samples {
			if v == 1 {
				data.sample_singletons = plusOneUpdateIntMap(data.sample_singletons, k)
			}
			if v > 1 && obc_ok && obiclean[k] == "i" {
				data.sample_obiclean_bad = plusOneUpdateIntMap(data.sample_obiclean_bad, k)
			}
		}
	} else if s.HasAttribute("sample") {
		sample, _ := s.GetStringAttribute("sample")
		data.samples = plusUpdateIntMap(data.samples, sample, s.Count())
		data.sample_variants = plusOneUpdateIntMap(data.sample_variants, sample)
		if s.Count() == 1 {
			data.sample_singletons = plusOneUpdateIntMap(data.sample_singletons, sample)
		}
	}

	if s.HasAttribute("obiclean_status") {
		data.has_obiclean_status++
	}

	if s.HasAttribute("obiclean_weight") {
		data.has_obiclean_weight++
	}

	for k, v := range s.Annotations() {
		switch {
		case obiutils.IsAMap(v):
			plusOneUpdateIntMap(data.map_tags, k)
		case obiutils.IsASlice(v):
			plusOneUpdateIntMap(data.vector_tags, k)
		default:
			plusOneUpdateIntMap(data.tags, k)
		}
	}

	return data
}

func ISummary(iterator obiiter.IBioSequence, summarise []string) map[string]interface{} {

	nproc := obioptions.CLIParallelWorkers()
	waiter := sync.WaitGroup{}

	summaries := make([]*DataSummary, nproc)

	for n := 0; n < nproc; n++ {
		for _, v := range summarise {
			summaries[n].map_summaries[v] = make(map[string]int, 0)
		}
	}

	ff := func(iseq obiiter.IBioSequence, summary *DataSummary) {

		for iseq.Next() {
			batch := iseq.Get()
			for _, seq := range batch.Slice() {
				summary.Update(seq)
			}
		}
		waiter.Done()
	}

	waiter.Add(nproc)

	summaries[0] = NewDataSummary()
	go ff(iterator, summaries[0])

	for i := 1; i < nproc; i++ {
		summaries[i] = NewDataSummary()
		go ff(iterator.Split(), summaries[i])
	}

	waiter.Wait()
	obiutils.WaitForLastPipe()

	rep := summaries[0]

	for i := 1; i < nproc; i++ {
		rep = rep.Add(summaries[i])
	}

	dict := make(map[string]interface{})

	dict["count"] = map[string]interface{}{
		"variants":     rep.variant_count,
		"reads":        rep.read_count,
		"total_length": rep.symbole_count,
	}

	if len(rep.tags)+len(rep.map_tags)+len(rep.vector_tags) > 0 {
		dict["annotations"] = map[string]interface{}{
			"scalar_attributes": len(rep.tags),
			"map_attributes":    len(rep.map_tags),
			"vector_attributes": len(rep.vector_tags),
			"keys":              make(map[string]map[string]int, 3),
		}

		if len(rep.tags) > 0 {
			((dict["annotations"].(map[string]interface{}))["keys"].(map[string]map[string]int))["scalar"] = rep.tags
		}

		if len(rep.map_tags) > 0 {
			((dict["annotations"].(map[string]interface{}))["keys"].(map[string]map[string]int))["map"] = rep.map_tags
		}

		if len(rep.vector_tags) > 0 {
			((dict["annotations"].(map[string]interface{}))["keys"].(map[string]map[string]int))["vector"] = rep.vector_tags
		}

		if len(rep.samples) > 0 {
			dict["samples"] = map[string]interface{}{
				"sample_count": len(rep.samples),
				"sample_stats": make(map[string]map[string]int, 2),
			}

			stats := ((dict["samples"].(map[string]interface{}))["sample_stats"].(map[string]map[string]int))
			for k, v := range rep.samples {
				stats[k] = map[string]int{
					"reads":      v,
					"variants":   rep.sample_variants[k],
					"singletons": rep.sample_singletons[k],
				}

				if rep.variant_count == rep.has_obiclean_status {
					stats[k]["obiclean_bad"] = rep.sample_obiclean_bad[k]
				}
			}
		}
	}
	return dict
}
