package obichunk

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

func ISequenceChunk(iterator obiiter.IBioSequence,
	classifier *obiseq.BioSequenceClassifier,
	onMemory bool,
	dereplicate bool,
	na string,
	statsOn obiseq.StatsOnDescriptions,
) (obiiter.IBioSequence, error) {

	if onMemory {
		return ISequenceChunkOnMemory(iterator, classifier)
	} else {
		return ISequenceChunkOnDisk(iterator, classifier, dereplicate, na, statsOn)
	}
}
