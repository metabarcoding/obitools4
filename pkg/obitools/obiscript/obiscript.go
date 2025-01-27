package obiscript

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obilua"
)

func CLIScriptPipeline() obiiter.Pipeable {

	pipe := obilua.LuaScriptPipe(CLIScriptFilename(), true, obidefault.ParallelWorkers())

	return pipe
}
