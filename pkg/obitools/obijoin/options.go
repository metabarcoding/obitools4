package obijoin

import (
	"strings"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _by = []string{}
var _join = ""
var _UpdateID = false
var _UpdateSequence = false
var _UpdateQuality = false

type By struct {
	Left  []string
	Right []string
}

func JoinOptionSet(options *getoptions.GetOpt) {

	options.StringSliceVar(&_by, "by", 1, 1,
		options.Alias("b"),
		options.Description("to declare join keys."))

	options.StringVar(&_join, "join-with", _join,
		options.Alias("j"),
		options.Description("file name of the file to join with."),
		options.Required("You must provide a file name to join with."))

	options.BoolVar(&_UpdateID, "update-id", _UpdateID,
		options.Alias("i"),
		options.Description("Update the sequence IDs in the joined file."))

	options.BoolVar(&_UpdateSequence, "update-sequence", _UpdateSequence,
		options.Alias("s"),
		options.Description("Update the sequence in the joined file."))

	options.BoolVar(&_UpdateQuality, "update-quality", _UpdateQuality,
		options.Alias("q"),
		options.Description("Update the quality in the joined file."))

}

// OptionSet adds to the basic option set every options declared for
// the obipcr command
func OptionSet(options *getoptions.GetOpt) {
	obiconvert.OptionSet(options)
	JoinOptionSet(options)
}

func CLIBy() By {
	if len(_by) == 0 {
		return By{
			Left:  []string{"id"},
			Right: []string{"id"},
		}
	}

	left := make([]string, len(_by))
	right := make([]string, len(_by))

	for i, v := range _by {
		vals := strings.Split(v, "=")
		left[i] = vals[0]
		right[i] = vals[0]
		if len(vals) > 1 {
			right[i] = vals[1]
		}
	}

	return By{Left: left, Right: right}
}

func CLIJoinWith() string {
	return _join
}

func CLIUpdateId() bool {
	return _UpdateID
}

func CLIUpdateSequence() bool {
	return _UpdateSequence
}

func CLIUpdateQuality() bool {
	return _UpdateQuality
}
