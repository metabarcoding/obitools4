package obidemerge

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

func MakeDemergeWorker(key string) obiseq.SeqWorker {
	desc := obiseq.MakeStatsOnDescription(key)
	f := func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {

		if sequence.HasStatsOn(key) {
			stats := sequence.StatsOn(desc, "NA")
			sequence.DeleteAttribute(obiseq.StatsOnSlotName(key))
			slice := obiseq.NewBioSequenceSlice(stats.Len())
			i := 0

			stats.RLock()
			defer stats.RUnlock()
			for k, v := range stats.Map() {
				(*slice)[i] = sequence.Copy()
				(*slice)[i].SetAttribute(key, k)
				(*slice)[i].SetCount(v)
				i++
			}

			return *slice, nil
		}

		return obiseq.BioSequenceSlice{sequence}, nil
	}

	return obiseq.SeqWorker(f)
}

func CLIDemergeSequences(iterator obiiter.IBioSequence) obiiter.IBioSequence {
	worker := MakeDemergeWorker(CLIDemergeSlot())
	return iterator.MakeIWorker(worker, false, obidefault.ParallelWorkers(), 0)
}
