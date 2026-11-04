package obiprimer

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/schollz/progressbar/v3"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
)

// stageBar is the progress bar of one stage of the search, counting the
// sequences (or any other unit) it has processed. A nil *stageBar does
// nothing, which is what newStageBar returns when no bar can be displayed
// (--no-progressbar, standard error not a terminal): the stages are logged
// anyway, see Search.
type stageBar struct {
	bar *progressbar.ProgressBar
}

// newStageBar starts a bar of n steps titled desc, or returns nil.
func newStageBar(desc string, n int) *stageBar {
	if n <= 0 || !obidefault.ProgressBarVisible() {
		return nil
	}
	return &stageBar{progressbar.NewOptions(n,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(15),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionSetDescription("["+desc+"]"),
		progressbar.OptionClearOnFinish(),
	)}
}

// Add advances the bar; it is safe for concurrent use.
func (b *stageBar) Add(n int) {
	if b != nil {
		_ = b.bar.Add(n)
	}
}

// Done ends the bar.
func (b *stageBar) Done() {
	if b != nil {
		_ = b.bar.Finish()
	}
}

// batchClock logs the progress of a stage made of batches, with an estimate of
// the time left, at most every few seconds: the batches of the amplicon passes
// can be thousands, each one a pass over the corpus.
type batchClock struct {
	what        string
	total       int
	start, last time.Time
}

func newBatchClock(what string, total int) *batchClock {
	now := time.Now()
	return &batchClock{what, total, now, now}
}

// Done reports that done batches out of the total are finished.
func (b *batchClock) Done(done int) {
	now := time.Now()
	if done < b.total && now.Sub(b.last) < 10*time.Second {
		return
	}
	b.last = now
	elapsed := now.Sub(b.start)
	remaining := time.Duration(float64(elapsed) / float64(done) * float64(b.total-done))
	log.Infof("%s: batch %d/%d, %v elapsed, about %v left", b.what, done, b.total,
		elapsed.Round(time.Second), remaining.Round(time.Second))
}
