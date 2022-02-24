package obiiter

import (
	"os"

	"github.com/schollz/progressbar/v3"
)

func (iterator IBioSequenceBatch) Speed() IBioSequenceBatch {
	newIter := MakeIBioSequenceBatch()

	newIter.Add(1)

	go func() {
		newIter.WaitAndClose()
	}()

	bar := progressbar.NewOptions(
		-1,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(15),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetDescription("[Sequence Processing]"))

	go func() {

		for iterator.Next() {
			batch := iterator.Get()
			l := batch.Length()
			newIter.Push(batch)
			bar.Add(l)
		}

		newIter.Done()
	}()

	return newIter
}


func SpeedPipe() Pipeable {
	f := func(iterator IBioSequenceBatch) IBioSequenceBatch {
		return iterator.Speed()
	}

	return f
}