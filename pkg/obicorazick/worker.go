package obicorazick

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"os"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"github.com/rrethy/ahocorasick"
	"github.com/schollz/progressbar/v3"
)

func AhoCorazickWorker(slot string, patterns []string) obiseq.SeqWorker {

	sizebatch:=10000000
	nmatcher := len(patterns) / sizebatch + 1
	log.Infof("Building AhoCorasick %d matcher for %d patterns in slot %s",
			 nmatcher, len(patterns), slot)

	if nmatcher == 0 {
		log.Errorln("No patterns provided")
	}

	matchers := make([]*ahocorasick.Matcher, nmatcher)
	ieme := make(chan int)
	mutex := &sync.WaitGroup{}
	npar := min(obidefault.ParallelWorkers(), nmatcher)
	mutex.Add(npar)

	var bar *progressbar.ProgressBar
	if obidefault.ProgressBar() {
		pbopt := make([]progressbar.Option, 0, 5)
		pbopt = append(pbopt,
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionSetWidth(15),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetDescription("Building AhoCorasick matcher..."),
		)

		bar = progressbar.NewOptions(nmatcher, pbopt...)
	}

	builder := func() {
		for i := range ieme  {
		matchers[i] = ahocorasick.CompileStrings(patterns[i*sizebatch:min((i+1)*sizebatch,len(patterns))])
		if bar != nil {
			bar.Add(1)
		}
		}
		mutex.Done()
	}

	for i := 0; i < npar; i++ {
		go builder()
	}

	for i := 0; i < nmatcher; i++ {
		ieme <- i
	}

	close(ieme)
	mutex.Wait()

	fslot := slot + "_Fwd"
	rslot := slot + "_Rev"

	f := func(s *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		matchesF := 0
		matchesR := 0
		b := s.Sequence()
		bc := s.ReverseComplement(false).Sequence()

		for _, matcher := range matchers {
			matchesF += len(matcher.FindAllByteSlice(b))
			matchesR += len(matcher.FindAllByteSlice(bc))
		}

		log.Debugln("Macthes = ", matchesF, matchesR)
		matches := matchesF + matchesR
		if matches > 0 {
			s.SetAttribute(slot, matches)
			s.SetAttribute(fslot, matchesF)
			s.SetAttribute(rslot, matchesR)
		}

		return obiseq.BioSequenceSlice{s}, nil
	}

	return f
}

func AhoCorazickPredicate(minMatches int, patterns []string) obiseq.SequencePredicate {

	matcher := ahocorasick.CompileStrings(patterns)

	f := func(s *obiseq.BioSequence) bool {
		matches := matcher.FindAllByteSlice(s.Sequence())
		return len(matches) >= minMatches
	}

	return f
}
