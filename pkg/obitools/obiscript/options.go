package obiscript

import (
	log "github.com/sirupsen/logrus"

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
		options.Alias("S"))

	options.BoolVar(&_askTemplate, "template", _askTemplate,
		options.Description("Print on the standard output a script template."),
	)
}

func OptionSet(options *getoptions.GetOpt) {
	ScriptOptionSet(options)
	obiconvert.OptionSet(false)(options)
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
	return `function begin()
    obicontext.item("compteur",0)
end

function worker(sequence)
    samples = sequence:attribute("merged_sample")
    samples["tutu"]=4
    sequence:attribute("merged_sample",samples)
    sequence:attribute("toto",44444)
    nb = obicontext.inc("compteur")
    sequence:id("seq_" .. nb)
    return sequence
end

function finish()
    print("compteur = " .. obicontext.item("compteur"))
end
	`
}
