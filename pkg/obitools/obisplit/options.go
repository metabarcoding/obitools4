package obisplit

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"slices"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiapat"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var _askTemplate = false
var _config = ""
var _pattern_error = 4
var _pattern_indel = false

func SplitOptionSet(options *getoptions.GetOpt) {

	options.StringVar(&_config, "config", _config,
		options.Description("The configuration file."),
		options.Alias("C"))

	options.BoolVar(&_askTemplate, "template", _askTemplate,
		options.Description("Print on the standard output a script template."),
	)

	options.IntVar(&_pattern_error, "pattern-error", _pattern_error,
		options.Description("Maximum number of allowed error during pattern matching"),
	)

	options.BoolVar(&_pattern_indel, "allows-indels", _pattern_indel,
		options.Description("Allows for indel during pattern matching"),
	)

}

func OptionSet(options *getoptions.GetOpt) {
	SplitOptionSet(options)
	obiconvert.OptionSet(options)
}

func CLIHasConfig() bool {
	return CLIConfigFile() != ""
}

func CLIConfigFile() string {
	return _config
}

func CLIConfig() []SplitSequence {
	// os.Open() opens specific file in
	// read-only mode and this return
	// a pointer of type os.File
	file, err := os.Open(CLIConfigFile())

	// Checks for the error
	if err != nil {
		log.Fatal("Error while reading the file", err)
	}

	// Closes the file
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()

	// Checks for the error
	if err != nil {
		fmt.Println("Error reading records")
	}

	config := make([]SplitSequence, 0, max(0, len(records)-1))

	header := records[0]

	pattern_idx := slices.Index(header, "T-tag")
	pool_idx := slices.Index(header, "pcr_pool")

	if pattern_idx == -1 {
		log.Fatalf("Config file %s doesn't contain `T-tag`column", CLIConfigFile())
	}

	if pool_idx == -1 {
		pool_idx = pattern_idx
	}

	// Loop to iterate through
	// and print each of the string slice
	for _, eachrecord := range records[1:] {

		fp, err := obiapat.MakeApatPattern(eachrecord[pattern_idx],
			CLIPatternError(), CLIPatternInDels())
		if err != nil {
			log.Fatalf("Error cannot compile pattern %s : %v",
				eachrecord[pattern_idx], err)
		}

		rv, err := fp.ReverseComplement()
		if err != nil {
			log.Fatalf("Error cannot reverse complement pattern %s: %v",
				eachrecord[pattern_idx], err)
		}

		config = append(config, SplitSequence{
			pattern:         eachrecord[pattern_idx],
			name:            eachrecord[pool_idx],
			forward_pattern: fp,
			reverse_pattern: rv,
		})
	}

	return config
}

func CLIPatternError() int {
	return _pattern_error
}

func CLIPatternInDels() bool {
	return _pattern_indel
}

func CLIAskConfigTemplate() bool {
	return _askTemplate
}

func CLIConfigTemplate() string {
	return `T-tag,pcr_pool
CGGCACCTGTTACGCAGCCACTATCGGCT,pool_1
CGGCAAGACCCTATTGCATTGGCGCGGCT,pool_2
`
}
