package obiutils

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

var globalLocker sync.WaitGroup
var globalLockerCounter = 0

// RegisterAPipe increments the global lock counter and adds a new pipe to the global wait group.
//
// No parameters.
// No return values.
func RegisterAPipe() {
	globalLocker.Add(1)
	globalLockerCounter++
	log.Debugln(globalLockerCounter, " Pipes are registered now")
}

// UnregisterPipe decrements the global lock counter and signals that a pipe has finished.
//
// No parameters.
// No return values.
func UnregisterPipe() {
	globalLocker.Done()
	globalLockerCounter--
	log.Debugln(globalLockerCounter, "are still registered")
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
