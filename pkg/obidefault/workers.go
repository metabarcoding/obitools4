package obidefault

import "runtime"

var _MaxAllowedCPU = runtime.NumCPU()
var _WorkerPerCore = 1.0

var _ReadWorkerPerCore = 0.25
var _WriteWorkerPerCore = 0.25

var _StrictReadWorker = 0
var _StrictWriteWorker = 0

var _ParallelFilesRead = 0

// CLIParallelWorkers returns the number of parallel workers used for
// computing the result.
//
// The number of parallel workers is determined by the command line option
// --max-cpu|-m and the environment variable OBIMAXCPU. This number is
// multiplied by the variable _WorkerPerCore.
//
// No parameters.
// Returns an integer representing the number of parallel workers.
func ParallelWorkers() int {
	return int(float64(MaxCPU()) * float64(WorkerPerCore()))
}

// CLIMaxCPU returns the maximum number of CPU cores allowed.
//
// The maximum number of CPU cores is determined by the command line option
// --max-cpu|-m and the environment variable OBIMAXCPU.
//
// No parameters.
// Returns an integer representing the maximum number of CPU cores allowed.
func MaxCPU() int {
	return _MaxAllowedCPU
}

func MaxCPUPtr() *int {
	return &_MaxAllowedCPU
}

// WorkerPerCore returns the number of workers per CPU core.
//
// No parameters.
// Returns a float64 representing the number of workers per CPU core.
func WorkerPerCore() float64 {
	return _WorkerPerCore
}

// SetWorkerPerCore sets the number of workers per CPU core.
//
// It takes a float64 parameter representing the number of workers
// per CPU core and does not return any value.
func SetWorkerPerCore(n float64) {
	_WorkerPerCore = n
}

// SetMaxCPU sets the maximum number of CPU cores allowed.
//
// n - an integer representing the new maximum number of CPU cores.
func SetMaxCPU(n int) {
	_MaxAllowedCPU = n
}

// SetReadWorker sets the number of workers for reading files.
//
// The number of worker dedicated to reading files is determined
// as the number of allowed CPU cores multiplied by number of read workers per core.
// Setting the number of read workers using this function allows to decouple the number
// of read workers from the number of CPU cores.
//
// n - an integer representing the number of workers to be set.
func SetStrictReadWorker(n int) {
	_StrictReadWorker = n
}

func SetStrictWriteWorker(n int) {
	_StrictWriteWorker = n
}

// SetReadWorkerPerCore sets the number of worker per CPU
// core for reading files.
//
// n float64
func SetReadWorkerPerCore(n float64) {
	_ReadWorkerPerCore = n
}

func SetWriteWorkerPerCore(n float64) {
	_WriteWorkerPerCore = n
}

// ReadWorker returns the number of workers for reading files.
//
// No parameters.
// Returns an integer representing the number of workers.
func StrictReadWorker() int {
	return _StrictReadWorker
}

func StrictWriteWorker() int {
	return _StrictWriteWorker
}

// CLIReadParallelWorkers returns the number of parallel workers used for
// reading files.
//
// The number of parallel workers is determined by the command line option
// --max-cpu|-m and the environment variable OBIMAXCPU. This number is
// multiplied by the variable _ReadWorkerPerCore.
//
// No parameters.
// Returns an integer representing the number of parallel workers.
func ReadParallelWorkers() int {
	if StrictReadWorker() == 0 {
		n := int(float64(MaxCPU()) * ReadWorkerPerCore())
		if n == 0 {
			n = 1
		}
		return n
	} else {
		return StrictReadWorker()
	}
}

func WriteParallelWorkers() int {
	if StrictWriteWorker() == 0 {
		n := int(float64(MaxCPU()) * WriteWorkerPerCore())
		if n == 0 {
			n = 1
		}
		return n
	} else {
		return StrictReadWorker()
	}
}

// ReadWorkerPerCore returns the number of worker per CPU core for
// computing the result.
//
// No parameters.
// Returns a float64 representing the number of worker per CPU core.
func ReadWorkerPerCore() float64 {
	return _ReadWorkerPerCore
}

func WriteWorkerPerCore() float64 {
	return _ReadWorkerPerCore
}

// ParallelFilesRead returns the number of files to be read in parallel.
//
// No parameters.
// Returns an integer representing the number of files to be read.
func ParallelFilesRead() int {
	if _ParallelFilesRead == 0 {
		return ReadParallelWorkers()
	} else {
		return _ParallelFilesRead
	}
}

// SetParallelFilesRead sets the number of files to be read in parallel.
//
// n - an integer representing the number of files to be set.
func SetParallelFilesRead(n int) {
	_ParallelFilesRead = n
}
