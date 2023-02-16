package obidistribute

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _FilenamePattern = ""
var _SequenceClassifierTag = ""
var _BatchCount = 0
var _HashSize = 0
var _NAValue = "NA"
var _append = false

func DistributeOptionSet(options *getoptions.GetOpt) {
	options.StringVar(&_FilenamePattern, "pattern", _FilenamePattern,
		options.Alias("p"),
		options.Required("You must provide at pattern for the file names "),
		options.Description("The template used to build the names of the output files. "+
			"The variable part is represented by '%s'. "+
			"Example : toto_%s.fastq."))

	options.StringVar(&_SequenceClassifierTag, "classifier", _SequenceClassifierTag,
		options.Alias("c"),
		options.Description("The name of a tag annotating thes sequences. "+
			"The name must corresponds to a string, a integer or a boolean value. "+
			"That value will be used to dispatch sequences amoong the different files"))

	options.StringVar(&_NAValue, "na-value", _NAValue,
		options.Description("Value used when the classifier tag is not defined for a sequence."))

	options.IntVar(&_BatchCount, "batches", 0,
		options.Alias("n"),
		options.Description("Indicates in how many batches the input file must bee splitted."))

	options.BoolVar(&_append, "append", _append,
		options.Alias("A"),
		options.Description("Indicates to append sequence to files if they already exist."))

	options.IntVar(&_HashSize, "hash", 0,
		options.Alias("H"),
		options.Description("Indicates to split the input into at most <n> batch based on a hash code of the seequence."))
}

func OptionSet(options *getoptions.GetOpt) {
	obiconvert.InputOptionSet(options)
	obiconvert.OutputOptionSet(options)
	DistributeOptionSet(options)
}

func CLIAppendSequences() bool {
	return _append
}

func CLISequenceClassifier() *obiseq.BioSequenceClassifier {
	switch {
	case _SequenceClassifierTag != "":
		return obiseq.AnnotationClassifier(_SequenceClassifierTag, _NAValue)
	case _BatchCount > 0:
		return obiseq.RotateClassifier(_BatchCount)
	case _HashSize > 0:
		return obiseq.HashClassifier(_HashSize)
	}

	log.Fatal("one of the options --classifier, -- hash or --batch must be specified")
	return nil
}

func CLIFileNamePattern() string {
	x := fmt.Sprintf(_FilenamePattern, "_xxx_")
	if strings.Contains(x, "(string=_xxx_)") {
		log.Panicf("patern %s is not correct : %s", _FilenamePattern, x)
	}

	return _FilenamePattern
}

func CLINAValue() string {
	return _NAValue
}
