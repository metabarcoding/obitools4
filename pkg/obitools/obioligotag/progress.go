package obioligotag

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
)

// cliqueProgress displays the progress of the search of a family: the bar goes
// from 0 to the required family size and follows the size of the best family
// found so far. A nil *cliqueProgress does nothing, which is what is used when
// no progress bar has to be displayed.
type cliqueProgress struct {
	bar  *progressbar.ProgressBar
	max  int
	stop chan struct{}
	done sync.WaitGroup
}

// startCliqueProgress starts the progress bar of a search for a family of
// target oligonucleotides, or returns nil if no bar can be displayed.
func startCliqueProgress(target int) *cliqueProgress {
	if !obidefault.ProgressBarVisible() {
		return nil
	}

	p := &cliqueProgress{
		max:  target,
		stop: make(chan struct{}),
		bar: progressbar.NewOptions(target,
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionSetWidth(15),
			progressbar.OptionShowCount(),
			progressbar.OptionSetElapsedTime(true),
			progressbar.OptionSetPredictTime(false),
			progressbar.OptionThrottle(100*time.Millisecond),
			progressbar.OptionSetDescription("[Searching a family of oligonucleotides]"),
		),
	}

	// The search may stay a long time on the same family: keep the elapsed
	// time alive.
	p.done.Add(1)
	go func() {
		defer p.done.Done()
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				p.bar.Add(0)
			case <-p.stop:
				return
			}
		}
	}()

	return p
}

// Update reports the size of the best family found so far.
func (p *cliqueProgress) Update(best int) {
	if p != nil {
		p.bar.Set(min(best, p.max))
	}
}

// Stop ends the display. The bar is filled only if the family was found.
func (p *cliqueProgress) Stop(reached bool) {
	if p == nil {
		return
	}
	close(p.stop)
	p.done.Wait()
	if reached {
		p.bar.Finish()
	} else {
		p.bar.Exit()
	}
	fmt.Fprintln(os.Stderr)
}
