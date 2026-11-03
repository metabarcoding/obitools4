package obiutils

import (
	"sync"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
)

var globalLocker sync.WaitGroup
var globalLockerCounter atomic.Int64

// RegisterAPipe increments the global lock counter and adds a new pipe to the global wait group.
//
// No parameters.
// No return values.
func RegisterAPipe() {
	globalLocker.Add(1)
	n := globalLockerCounter.Add(1)
	log.Debugln(n, " Pipes are registered now")
}

// UnregisterPipe decrements the global lock counter and signals that a pipe has finished.
//
// No parameters.
// No return values.
func UnregisterPipe() {
	globalLocker.Done()
	n := globalLockerCounter.Add(-1)
	log.Debugln(n, "are still registered")
}

// WaitForLastPipe waits until all registered pipes have finished.
//
// THe function have to be called at the end of every main function.
//
// No parameters.
// No return values.
func WaitForLastPipe() {
	globalLocker.Wait()
}
