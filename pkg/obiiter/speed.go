package obiiter

import (
	"fmt"
	"os"

	"github.com/schollz/progressbar/v3"
)

func (iterator IBioSequence) Speed(message ...string) IBioSequence {

	// If the STDERR is redicted and doesn't end up to a terminal
	// No progress bar is printed.
	o, _ := os.Stderr.Stat()
	if (o.Mode() & os.ModeCharDevice) != os.ModeCharDevice {
		return iterator
	}

	newIter := MakeIBioSequence()

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

		fmt.Fprintln(os.Stderr)
		newIter.Done()
	}()

	return newIter
}

func SpeedPipe(message ...string) Pipeable {
	f := func(iterator IBioSequence) IBioSequence {
		return iterator.Speed(message...)
	}

	return f
}
