package obiiter

import (
	"fmt"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

func (iterator IBioSequence) Speed(message string, size ...int) IBioSequence {

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
		progressbar.OptionSetDescription(message),
	)

	n := -1
	if len(size) > 0 {
		n = size[0]
	}

	bar := progressbar.NewOptions(n, pbopt...)

	go func() {
		c := 0
		start := time.Now()

		for iterator.Next() {
			batch := iterator.Get()
			c += batch.Len()
			newIter.Push(batch)
			elapsed := time.Since(start)
			if elapsed > (time.Millisecond * 100) {
				bar.Add(c)
				c = 0
				start = time.Now()
			}
		}

		fmt.Fprintln(os.Stderr)
		newIter.Done()
	}()

	if iterator.IsPaired() {
		newIter.MarkAsPaired()
	}

	return newIter
}

func SpeedPipe(message string) Pipeable {
	f := func(iterator IBioSequence) IBioSequence {
		return iterator.Speed(message)
	}

	return f
}
