package obiiter

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

func IFragments(minsize, length, overlap, size, nworkers int) Pipeable {
	step := length - overlap

	ifrg := func(iterator IBioSequence) IBioSequence {
		newiter := MakeIBioSequence()
		iterator = iterator.SortBatches()

		newiter.Add(nworkers)

		go func() {
			newiter.WaitAndClose()
		}()

		f := func(iterator IBioSequence) {
			source := ""
			for iterator.Next() {
				news := obiseq.MakeBioSequenceSlice()
				sl := iterator.Get()
				source = sl.Source()
				for _, s := range sl.Slice() {

					if s.Len() <= minsize {
						news = append(news, s)
					} else {
						for i := 0; i < s.Len(); i += step {
							end := min(i+length, s.Len())
							fusion := false
							if (s.Len() - end) < step {
								end = s.Len()
								fusion = true
							}
							frg, err := s.Subsequence(i, end, false)

							if err != nil {
								log.Panicln(err)
							}
							news = append(news, frg)
							// if len(news) >= size {
							// 	newiter.Push(MakeBioSequenceBatch(order(), news))
							// 	news = obiseq.MakeBioSequenceSlice()
							// }
							if fusion {
								i = s.Len()
							}
						}
						s.Recycle()
					}
				} // End of the slice loop
				newiter.Push(MakeBioSequenceBatch(source, sl.Order(), news))
			} // End of the iterator loop

			// if len(news) > 0 {
			// 	newiter.Push(MakeBioSequenceBatch(order(), news))
			// }

			newiter.Done()

		}

		for i := 1; i < nworkers; i++ {
			go f(iterator.Split())
		}
		go f(iterator)

		return newiter.SortBatches().Rebatch(size)
	}

	return ifrg
}
