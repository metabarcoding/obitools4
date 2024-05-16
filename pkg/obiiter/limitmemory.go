package obiiter

import (
	"runtime"

	"github.com/pbnjay/memory"
	log "github.com/sirupsen/logrus"
)

func (iterator IBioSequence) LimitMemory(fraction float64) IBioSequence {
	newIter := MakeIBioSequence()

	fracLoad := func() float64 {
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		return float64(mem.Alloc) / float64(memory.TotalMemory())
	}

	newIter.Add(1)
	go func() {

		for iterator.Next() {
			nwait := 0
			for fracLoad() > fraction {
				runtime.Gosched()
				nwait++
				if nwait%1000 == 0 {
					log.Warnf("Wait for memory limit %f/%f", fracLoad(), fraction)

				}
				if nwait > 10000 {
					log.Warnf("Very long wait for memory limit %f/%f", fracLoad(), fraction)
					break
				}
			}
			newIter.Push(iterator.Get())
		}

		newIter.Done()
	}()

	go func() {
		newIter.WaitAndClose()
	}()

	return newIter
}
