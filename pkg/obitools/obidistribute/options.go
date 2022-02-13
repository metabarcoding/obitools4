package obidistribute

import (
	"fmt"
	"log"
	"strings"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _FilenamePattern = ""
var _SequenceClassifierTag = ""
var _BatchCount = 0

func DistributeOptionSet(options *getoptions.GetOpt) {
	options.StringVar(&_FilenamePattern, "pattern", _FilenamePattern,
		options.Alias("p"),
		options.Required("You must provide at pattern for the file names "),
		options.Description("The N first sequence records of the file are discarded from the analysis and not reported to the output file."))

	options.StringVar(&_SequenceClassifierTag, "classifier", _SequenceClassifierTag,
		options.Alias("c"),
		options.Description("The N first sequence records of the file are discarded from the analysis and not reported to the output file."))

	options.IntVar(&_BatchCount, "batch", 0,
		options.Alias("n"),
		options.Description("The N first sequence records of the file are discarded from the analysis and not reported to the output file."))
}

func OptionSet(options *getoptions.GetOpt) {
	obiconvert.InputOptionSet(options)
	obiconvert.OutputOptionSet(options)
	DistributeOptionSet(options)
}

func CLISequenceClassifier() obiseq.SequenceClassifier {
	switch {
	case _SequenceClassifierTag != "":
		return obiseq.AnnotationClassifier(_SequenceClassifierTag)
	case _BatchCount > 0:
		return obiseq.RotateClassifier(_BatchCount)

	}

	log.Fatal("one of the options --classifier or --batch must be specified")
	return nil
}

func CLIFileNamePattern() string {
	x := fmt.Sprintf(_FilenamePattern, "_xxx_")
	if strings.Contains(x, "(string=_xxx_)") {
		log.Panicf("patern %s is not correct : %s", _FilenamePattern, x)
	}

	return _FilenamePattern
}
