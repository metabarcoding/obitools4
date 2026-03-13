package obidefault

var _BatchSize = 2000

// SetBatchSize sets the size of the sequence batches.
//
// n - an integer representing the size of the sequence batches.
func SetBatchSize(n int) {
	_BatchSize = n
}

// CLIBatchSize returns the expected size of the sequence batches.
//
// In Obitools, the sequences are processed in parallel by batches.
// The number of sequence in each batch is determined by the command line option
// --batch-size and the environment variable OBIBATCHSIZE.
//
// No parameters.
// Returns an integer value.
func BatchSize() int {
	return _BatchSize
}

func BatchSizePtr() *int {
	return &_BatchSize
}

// _BatchMem holds the maximum cumulative memory (in bytes) per batch when
// memory-based batching is requested. A value of 0 disables memory-based
// batching and falls back to count-based batching.
var _BatchMem = 0
var _BatchMemStr = ""

// SetBatchMem sets the memory budget per batch in bytes.
func SetBatchMem(n int) {
	_BatchMem = n
}

// BatchMem returns the current memory budget per batch in bytes.
// A value of 0 means memory-based batching is disabled.
func BatchMem() int {
	return _BatchMem
}

func BatchMemPtr() *int {
	return &_BatchMem
}

// BatchMemStr returns the raw --batch-mem string value as provided on the CLI.
func BatchMemStr() string {
	return _BatchMemStr
}

func BatchMemStrPtr() *string {
	return &_BatchMemStr
}
