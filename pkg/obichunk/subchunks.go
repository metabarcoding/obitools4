package obichunk

import (
	"sort"
	"sync/atomic"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

//
// Interface for sorting a list of sequences accoording to
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

func ISequenceSubChunk(iterator obiiter.IBioSequenceBatch,
	classifier *obiseq.BioSequenceClassifier,
	sizes ...int) (obiiter.IBioSequenceBatch, error) {

	bufferSize := iterator.BufferSize()
	nworkers := 4

	if len(sizes) > 0 {
		nworkers = sizes[0]
	}

	if len(sizes) > 1 {
		bufferSize = sizes[1]
	}

	newIter := obiiter.MakeIBioSequenceBatch(bufferSize)

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

	ff := func(iterator obiiter.IBioSequenceBatch,
		classifier *obiseq.BioSequenceClassifier) {

		ordered := make([]sSS, 100)

		for iterator.Next() {

			batch := iterator.Get()

			if batch.Length() > 1 {
				classifier.Reset()

				if cap(ordered) < batch.Length() {
					log.Debugln("Allocate a new ordered sequences : ", batch.Length())
					ordered = make([]sSS, batch.Length())
				} else {
					ordered = ordered[:batch.Length()]
				}

				for i, s := range batch.Slice() {
					ordered[i].code = classifier.Code(s)
					ordered[i].seq = s
					batch.Slice()[i] = nil
				}

				batch.Recycle()

				_By(func(p1, p2 *sSS) bool {
					return p1.code < p2.code
				}).Sort(ordered)

				last := ordered[0].code
				ss := obiseq.MakeBioSequenceSlice()
				for i, v := range ordered {
					if v.code != last {
						newIter.Push(obiiter.MakeBioSequenceBatch(nextOrder(), ss))
						ss = obiseq.MakeBioSequenceSlice()
						last = v.code
					}

					ss = append(ss, v.seq)
					ordered[i].seq = nil
				}

				if len(ss) > 0 {
					newIter.Push(obiiter.MakeBioSequenceBatch(nextOrder(), ss))
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