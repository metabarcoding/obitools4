package obiiter

import (
	log "github.com/sirupsen/logrus"
)

func (b BioSequenceBatch) IsPaired() bool {
	return b.slice.IsPaired()
}

func (b BioSequenceBatch) PairedWith() BioSequenceBatch {
	return MakeBioSequenceBatch(
		b.Source(),
		b.order,
		*b.slice.PairedWith(),
	)
}

func (b *BioSequenceBatch) PairTo(p *BioSequenceBatch) {

	if b.order != p.order {
		log.Fatalf("both batches are not synchronized : (%d,%d)",
			b.order, p.order,
		)
	}

	b.slice.PairTo(&p.slice)

}

func (b *BioSequenceBatch) UnPair() {
	b.slice.UnPair()
}

func (iter IBioSequence) MarkAsPaired() {
	iter.pointer.paired = true
}

func (iter IBioSequence) PairTo(p IBioSequence) IBioSequence {

	newIter := MakeIBioSequence()

	iter = iter.SortBatches()
	p = p.SortBatches()

	newIter.Add(1)

	go func() {
		newIter.WaitAndClose()
	}()

	go func() {

		for iter.Next() {
			p.Next()
			batch := iter.Get()
			pbatch := p.Get()
			batch.PairTo(&pbatch)
			newIter.Push(batch)
		}

		newIter.Done()
	}()

	newIter.MarkAsPaired()
	return newIter

}

func (iter IBioSequence) PairedWith() IBioSequence {

	newIter := MakeIBioSequence()

	newIter.Add(1)

	go func() {
		newIter.WaitAndClose()
	}()

	go func() {

		for iter.Next() {
			batch := iter.Get().PairedWith()
			newIter.Push(batch)
		}

		newIter.Done()
	}()

	newIter.MarkAsPaired()
	return newIter

}

func (iter IBioSequence) IsPaired() bool {
	return iter.pointer.paired
}
