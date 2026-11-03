package obiprimers

import (
	"os"
	"runtime/debug"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
)

// gcPercent is the garbage collector target of the command: the heap may grow
// by that percentage over the live data before the next collection (the Go
// default is 100). Reading the sequences and the search allocate large buffers
// and tables that die as soon as a stage is over, and with the default the
// process holds up to twice what it needs at its peak. Peak of memory and CPU
// time against the default (see OBIPRIMERS_NOTES.md, section 26): 7822
// vertebrate genomes, 50 gives -8% and +0%, 25 -13% and +8%; 18061 genomes
// (593 Mb), 50 gives -20% and +3%, 25 -29% and +8.5%.
const gcPercent = 50

// batchSizeMax is the default of the --batch-size-max option for this command.
// The reading goroutines hand the sequences over by batches, and a batch of 2000
// genomes of 16 kb (the general default) is 33 Mb of bases once parsed, times
// the number of workers and the batches waiting between the stages. Genomes are
// few and long: 100 per batch keeps all the cores busy.
const batchSizeMax = 100

// SetReadingDefaults gives the options shared by the OBITools their defaults for
// this command; it must be called before the options are parsed, which still
// override them.
func SetReadingDefaults() {
	*obidefault.BatchSizeMaxPtr() = batchSizeMax
}

// tuneMemory sets the collector target. A user who set GOGC keeps it.
func tuneMemory() {
	if os.Getenv("GOGC") == "" {
		debug.SetGCPercent(gcPercent)
	}
}

// releaseMemory collects the garbage and gives the free memory back to the
// system. Without it the memory freed by a stage stays in the process (the
// runtime returns it slowly), and the next stage takes new memory on top of it:
// the peak is then the sum of the stages instead of the largest.
func releaseMemory() { debug.FreeOSMemory() }
