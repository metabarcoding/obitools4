package obiscript

import (
	"log"
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obigrep"
	"github.com/DavidGamba/go-getoptions"
)

var _script = ""
var _askTemplate = false

func ScriptOptionSet(options *getoptions.GetOpt) {

	options.StringVar(&_script, "script", _script,
		options.Description("The script to execute."),
		options.Alias("S"),
		options.Description("Name of a map attribute."))

	options.BoolVar(&_askTemplate, "template", _askTemplate,
		options.Description("Print on the standard output a script template."),
	)
}

func OptionSet(options *getoptions.GetOpt) {
	ScriptOptionSet(options)
	obiconvert.OptionSet(options)
	obigrep.SequenceSelectionOptionSet(options)
}

func CLIScriptFilename() string {
	return _script
}

func CLIScript() string {
	file, err := os.ReadFile(_script) // Reads the script
	if err != nil {
		log.Fatalf("cannot read the script file : %s", _script)
	}
	return string(file)
}

func CLIAskScriptTemplate() bool {
	return _askTemplate
}

func CLIScriptTemplate() string {
	return `
	import {
		"sync"
		"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	}
	//
	// Begin function run before the first sequence being processed
	//

	func Begin(environment *sync.Map) {

	}

	//
	// Begin function run after the last sequence being processed
	//

	func End(environment *sync.Map) {

	}

	//
	// Worker function run for each sequence validating the selection predicat as specified by 
	// the command line options.
	//
	// The function must return the altered sequence.
	// If the function returns nil, the sequence is discarded from the output
	func Worker(sequence *obiseq.BioSequence, environment *sync.Map) *obiseq.BioSequence {


		return sequence
	}
	`
}
