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
