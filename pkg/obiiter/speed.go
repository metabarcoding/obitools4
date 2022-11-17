package obiiter

import (
	"os"

	"github.com/schollz/progressbar/v3"
)

func (iterator IBioSequenceBatch) Speed(message ...string) IBioSequenceBatch {
	newIter := MakeIBioSequenceBatch()

	newIter.Add(1)

	go func() {
		newIter.WaitAndClose()
	}()

	pbopt := make([]progressbar.Option, 0, 5)
	pbopt = append(pbopt,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(15),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
	)

	if len(message) > 0 {
		pbopt = append(pbopt,
			progressbar.OptionSetDescription(message[0]),
		)
	} else {
		pbopt = append(pbopt,
			progressbar.OptionSetDescription("[Sequence Processing]"),
		)
	}

	bar := progressbar.NewOptions(-1, pbopt...)

	go func() {

		for iterator.Next() {
			batch := iterator.Get()
			l := batch.Len()
			newIter.Push(batch)
			bar.Add(l)
		}

		newIter.Done()
	}()

	return newIter
}

func SpeedPipe(message ...string) Pipeable {
	f := func(iterator IBioSequenceBatch) IBioSequenceBatch {
		return iterator.Speed(message...)
	}

	return f
}
