package obilua

import (
	"bytes"
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

func NewInterpreter() *lua.LState {
	lua := lua.NewState()

	RegisterObilib(lua)

	return lua
}

func Compile(program []byte, name string) (*lua.FunctionProto, error) {

	reader := bytes.NewReader(program)
	chunk, err := parse.Parse(reader, name)
	if err != nil {
		return nil, err
	}

	proto, err := lua.Compile(chunk, name)
	if err != nil {
		return nil, err
	}
	return proto, nil
}

func CompileScript(filePath string) (*lua.FunctionProto, error) {
	program, err := os.ReadFile(filePath)

	if err != nil {
		return nil, err
	}

	return Compile(program, filePath)
}

func LuaWorker(proto *lua.FunctionProto) obiseq.SeqWorker {
	interpreter := NewInterpreter()
	lfunc := interpreter.NewFunctionFromProto(proto)

	f := func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		interpreter.SetGlobal("sequence", obiseq2Lua(interpreter, sequence))
		interpreter.Push(lfunc)
		err := interpreter.PCall(0, lua.MultRet, nil)

		return obiseq.BioSequenceSlice{sequence}, err
	}

	return f
}

func LuaProcessor(iterator obiiter.IBioSequence, name, program string, breakOnError bool, nworkers int) obiiter.IBioSequence {
	newIter := obiiter.MakeIBioSequence()

	if nworkers <= 0 {
		nworkers = obioptions.CLIParallelWorkers()
	}

	newIter.Add(nworkers)

	go func() {
		newIter.WaitAndClose()
	}()

	bp := []byte(program)
	proto, err := Compile(bp, name)

	if err != nil {
		log.Fatalf("Cannot compile script %s : %v", name, err)
	}

	ff := func(iterator obiiter.IBioSequence) {
		w := LuaWorker(proto)
		sw := obiseq.SeqToSliceWorker(w, false)

		// iterator = iterator.SortBatches()

		for iterator.Next() {
			seqs := iterator.Get()
			slice := seqs.Slice()
			ns, err := sw(slice)

			if err != nil {
				if breakOnError {
					log.Fatalf("Error during Lua sequence processing : %v", err)
				} else {
					log.Warnf("Error during Lua sequence processing : %v", err)
				}
			}

			newIter.Push(obiiter.MakeBioSequenceBatch(seqs.Order(), ns))
			seqs.Recycle(false)
		}

		newIter.Done()
	}

	for i := 1; i < nworkers; i++ {
		go ff(iterator.Split())
	}

	go ff(iterator)

	if iterator.IsPaired() {
		newIter.MarkAsPaired()
	}

	return newIter

}

func LuaPipe(name, program string, breakOnError bool, nworkers int) obiiter.Pipeable {

	f := func(input obiiter.IBioSequence) obiiter.IBioSequence {
		return LuaProcessor(input, name, program, breakOnError, nworkers)
	}

	return f
}

func LuaScriptPipe(filename string, breakOnError bool, nworkers int) obiiter.Pipeable {
	program, err := os.ReadFile(filename)

	if err != nil {
		log.Fatalf("Cannot read script file %s", filename)
	}

	return LuaPipe(filename, string(program), breakOnError, nworkers)
}
