package obijoin

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiformats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"

	log "github.com/sirupsen/logrus"
)

type IndexedSequenceSlice struct {
	Sequences obiseq.BioSequenceSlice
	Indices   []map[interface{}]*obiutils.Set[int]
}

func (s IndexedSequenceSlice) Len() int {
	return len(s.Sequences)
}

func (s IndexedSequenceSlice) Get(keys ...interface{}) *obiseq.BioSequenceSlice {
	var keeps obiutils.Set[int]

	for i, v := range s.Indices {
		if i == 0 {
			keeps = *v[keys[0]]
		} else {
			keeps = keeps.Intersection(*v[keys[i]])
		}
	}

	rep := obiseq.MakeBioSequenceSlice(len(keeps))
	for i, v := range keeps.Members() {
		rep[i] = s.Sequences[v]
	}

	return &rep
}

func BuildIndexedSequenceSlice(seqs obiseq.BioSequenceSlice, keys []string) IndexedSequenceSlice {
	indices := make([]map[interface{}]*obiutils.Set[int], len(keys))

	for i, k := range keys {
		idx := make(map[interface{}]*obiutils.Set[int])

		for j, seq := range seqs {

			if value, ok := seq.GetAttribute(k); ok {
				goods, ok := idx[value]
				if !ok {
					goods = obiutils.NewSet[int]()
					idx[value] = goods
				}

				goods.Add(j)
			}
		}

		indices[i] = idx
	}

	return IndexedSequenceSlice{seqs, indices}
}

func MakeJoinWorker(by []string, index IndexedSequenceSlice, updateId, updateSequence, updateQuality bool) obiseq.SeqWorker {
	f := func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		var ok bool

		keys := make([]interface{}, len(by))

		for i, v := range by {
			keys[i], ok = sequence.GetAttribute(v)
			if !ok {
				return obiseq.BioSequenceSlice{sequence}, nil
			}
		}

		join_with := index.Get(keys...)

		rep := obiseq.MakeBioSequenceSlice(join_with.Len())

		if join_with.Len() == 0 {
			return obiseq.BioSequenceSlice{sequence}, nil
		}

		for i, v := range *join_with {
			rep[i] = sequence.Copy()
			annot := rep[i].Annotations()
			new_annot := v.Annotations()

			for k, v := range new_annot {
				annot[k] = v
			}

			if updateId {
				rep[i].SetId(v.Id())
			}
			if updateSequence && len(v.Sequence()) > 0 {
				rep[i].SetSequence(v.Sequence())
			}
			if updateQuality && len(v.Qualities()) > 0 {
				rep[i].SetQualities(v.Qualities())
			}
		}

		return rep, nil
	}

	return obiseq.SeqWorker(f)
}

func CLIJoinSequences(iterator obiiter.IBioSequence) obiiter.IBioSequence {

	data_iter, err := obiformats.ReadSequencesFromFile(CLIJoinWith())

	if err != nil {
		log.Fatalf("Cannot read the data file to merge with: %s %v", CLIJoinWith(), err)
	}

	data := data_iter.Load()

	keys := CLIBy()

	index := BuildIndexedSequenceSlice(data, keys.Right)

	worker := MakeJoinWorker(keys.Left, index, CLIUpdateId(), CLIUpdateSequence(), CLIUpdateQuality())

	iterator = iterator.MakeIWorker(worker, false, obioptions.CLIParallelWorkers())

	return iterator
}
