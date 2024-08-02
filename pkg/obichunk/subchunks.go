package obichunk

import (
	"sort"
	"sync/atomic"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

//
// Interface for sorting a list of sequences according to
// their classes
//

type sSS struct {
	code int
	seq  *obiseq.BioSequence
}

// By is the type of a "less" function that defines the ordering of its Planet arguments.
type _By func(p1, p2 *sSS) bool

type sSSSorter struct {
	seqs []sSS
	by   _By // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *sSSSorter) Len() int {
	return len(s.seqs)
}

// Swap is part of sort.Interface.
func (s *sSSSorter) Swap(i, j int) {
	s.seqs[i], s.seqs[j] = s.seqs[j], s.seqs[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *sSSSorter) Less(i, j int) bool {
	return s.by(&s.seqs[i], &s.seqs[j])
}

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by _By) Sort(seqs []sSS) {
	ps := &sSSSorter{
		seqs: seqs,
		by:   by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ps)
}

//
// End of the sort interface
//

func ISequenceSubChunk(iterator obiiter.IBioSequence,
	classifier *obiseq.BioSequenceClassifier,
	nworkers int) (obiiter.IBioSequence, error) {

	if nworkers <= 0 {
		nworkers = obioptions.CLIParallelWorkers()
	}

	newIter := obiiter.MakeIBioSequence()

	newIter.Add(nworkers)

	go func() {
		newIter.Wait()
		newIter.Close()
	}()

	//omutex := sync.Mutex{}
	order := int32(0)

	nextOrder := func() int {
		neworder := int(atomic.AddInt32(&order, 1))
		return neworder
	}

	ff := func(iterator obiiter.IBioSequence,
		classifier *obiseq.BioSequenceClassifier) {

		ordered := make([]sSS, 100)

		for iterator.Next() {

			batch := iterator.Get()
			source := batch.Source()
			if batch.Len() > 1 {
				classifier.Reset()

				if cap(ordered) < batch.Len() {
					log.Debugln("Allocate a new ordered sequences : ", batch.Len())
					ordered = make([]sSS, batch.Len())
				} else {
					ordered = ordered[:batch.Len()]
				}

				for i, s := range batch.Slice() {
					ordered[i].code = classifier.Code(s)
					ordered[i].seq = s
					batch.Slice()[i] = nil
				}

				batch.Recycle(false)

				_By(func(p1, p2 *sSS) bool {
					return p1.code < p2.code
				}).Sort(ordered)

				last := ordered[0].code
				ss := obiseq.MakeBioSequenceSlice()
				for i, v := range ordered {
					if v.code != last {
						newIter.Push(obiiter.MakeBioSequenceBatch(source, nextOrder(), ss))
						ss = obiseq.MakeBioSequenceSlice()
						last = v.code
					}

					ss = append(ss, v.seq)
					ordered[i].seq = nil
				}

				if len(ss) > 0 {
					newIter.Push(obiiter.MakeBioSequenceBatch(source, nextOrder(), ss))
				}
			} else {
				newIter.Push(batch.Reorder(nextOrder()))
			}
		}

		newIter.Done()
	}

	for i := 0; i < nworkers-1; i++ {
		go ff(iterator.Split(), classifier.Clone())
	}
	go ff(iterator, classifier)

	return newIter, nil
}
