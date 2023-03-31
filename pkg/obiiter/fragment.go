package obiiter

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiutils"
)

func IFragments(minsize, length, overlap, size, nworkers int) Pipeable {
	step := length - overlap

	ifrg := func(iterator IBioSequence) IBioSequence {
		order := obiutils.AtomicCounter()
		newiter := MakeIBioSequence()
		iterator = iterator.SortBatches()

		newiter.Add(nworkers)

		go func() {
			newiter.WaitAndClose()
		}()

		f := func(iterator IBioSequence, id int) {
			news := obiseq.MakeBioSequenceSlice()
			for iterator.Next() {
				sl := iterator.Get()
				for _, s := range sl.Slice() {
					if s.Len() <= minsize {
						news = append(news, s)
					} else {
						for i := 0; i < s.Len(); i += step {
							end := obiutils.MinInt(i+length, s.Len())
							frg, err := s.Subsequence(i, end, false)
							if err != nil {
								log.Panicln(err)
							}
							news = append(news, frg)
							if len(news) >= size {
								newiter.Push(MakeBioSequenceBatch(order(), news))
								news = obiseq.MakeBioSequenceSlice()
							}
						}
						s.Recycle()
					}
					if len(news) >= size {
						o := order()
						newiter.Push(MakeBioSequenceBatch(o, news))
						news = obiseq.MakeBioSequenceSlice()
					}
				} // End of the slice loop
				sl.Recycle(false)
			} // End of the iterator loop

			if len(news) > 0 {
				newiter.Push(MakeBioSequenceBatch(order(), news))
			}

			newiter.Done()

		}

		for i := 1; i < nworkers; i++ {
			go f(iterator.Split(), i)
		}
		go f(iterator, 0)

		return newiter
	}

	return ifrg
}
