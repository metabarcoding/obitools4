package obiblackboard

import (
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
)

type SequenceCounter struct {
	Variants    int
	Reads       int
	Nucleotides int
	Runner      DoTask
}

func CountSequenceAggregator(target string) *SequenceCounter {
	cc := &SequenceCounter{
		Variants:    0,
		Reads:       0,
		Nucleotides: 0,
		Runner:      nil,
	}

	mutex := sync.Mutex{}

	runner := func(bb *Blackboard, task *Task) *Task {
		body := task.Body.(obiiter.BioSequenceBatch)

		mutex.Lock()
		cc.Variants += body.Len()
		cc.Reads += body.Slice().Count()
		cc.Nucleotides += body.Slice().Size()
		mutex.Unlock()

		nt := task.GetNext(target, true, false)
		return nt
	}

	cc.Runner = runner
	return cc
}

func RecycleSequences(rescycleSequence bool, target string) DoTask {
	return func(bb *Blackboard, task *Task) *Task {
		body := task.Body.(obiiter.BioSequenceBatch)
		// log.Warningf("With priority %d, Recycling %s[%d]", task.Priority, body.Source(), body.Order())
		body.Recycle(rescycleSequence)
		return task.GetNext(target, false, false)
	}
}
