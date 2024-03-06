package obiscript

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obilua"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
)

func CLIScriptPipeline() obiiter.Pipeable {

	pipe := obilua.LuaScriptPipe(CLIScriptFilename(), true, obioptions.CLIParallelWorkers())

	return pipe
}
