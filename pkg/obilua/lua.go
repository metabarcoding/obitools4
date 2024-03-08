package obilua

import (
	"bytes"
	"fmt"
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
	RegisterObiContext(lua)

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
	interpreter.Push(lfunc)
	err := interpreter.PCall(0, lua.MultRet, nil)

	if err != nil {
		log.Fatalf("Error in executing the lua script")
	}

	result := interpreter.GetGlobal("worker")

	if lua_worker, ok := result.(*lua.LFunction); ok {
		f := func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
			// Call the Lua function concat
			// Lua analogue:
			//	str = concat("Go", "Lua")
			//	print(str)
			if err := interpreter.CallByParam(lua.P{
				Fn:      lua_worker, // name of Lua function
				NRet:    1,          // number of returned values
				Protect: true,       // return err or panic
			}, obiseq2Lua(interpreter, sequence)); err != nil {
				log.Fatal(err)
			}

			lreponse := interpreter.Get(-1)
			defer interpreter.Pop(1)

			if reponse, ok := lreponse.(*lua.LUserData); ok {
				s := reponse.Value
				switch val := s.(type) {
				case *obiseq.BioSequence:
					return obiseq.BioSequenceSlice{val}, err
				default:
					return nil, fmt.Errorf("worker function doesn't return the correct type")
				}
			}

			return nil, fmt.Errorf("worker function doesn't return the correct type")
		}

		return f
	}

	log.Fatalf("THe worker object is not a function")
	return nil
	// f := func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
	// 	interpreter.SetGlobal("sequence", obiseq2Lua(interpreter, sequence))
	// 	interpreter.Push(lfunc)
	// 	err := interpreter.PCall(0, lua.MultRet, nil)
	// 	result := interpreter.GetGlobal("result")

	// 	if result != lua.LNil {
	// 		log.Info("youpi ", result)
	// 	}

	// 	rep := interpreter.GetGlobal("sequence")

	// 	if rep.Type() == lua.LTUserData {
	// 		ud := rep.(*lua.LUserData)
	// 		sequence = ud.Value.(*obiseq.BioSequence)
	// 	}

	// 	return obiseq.BioSequenceSlice{sequence}, err
	// }

}

func LuaProcessor(iterator obiiter.IBioSequence, name, program string, breakOnError bool, nworkers int) obiiter.IBioSequence {
	newIter := obiiter.MakeIBioSequence()

	if nworkers <= 0 {
		nworkers = obioptions.CLIParallelWorkers()
	}

	newIter.Add(nworkers)

	bp := []byte(program)
	proto, err := Compile(bp, name)

	if err != nil {
		log.Fatalf("Cannot compile script %s : %v", name, err)
	}

	interpreter := NewInterpreter()
	lfunc := interpreter.NewFunctionFromProto(proto)
	interpreter.Push(lfunc)
	err = interpreter.PCall(0, lua.MultRet, nil)

	if err != nil {
		log.Fatalf("Error in executing the lua script")
	}

	result := interpreter.GetGlobal("begin")
	if lua_begin, ok := result.(*lua.LFunction); ok {
		if err := interpreter.CallByParam(lua.P{
			Fn:      lua_begin, // name of Lua function
			NRet:    0,         // number of returned values
			Protect: true,      // return err or panic
		}); err != nil {
			log.Fatal(err)
		}
	}

	interpreter.Close()

	go func() {
		newIter.WaitAndClose()

		interpreter := NewInterpreter()
		lfunc := interpreter.NewFunctionFromProto(proto)
		interpreter.Push(lfunc)
		err = interpreter.PCall(0, lua.MultRet, nil)

		if err != nil {
			log.Fatalf("Error in executing the lua script")
		}

		result := interpreter.GetGlobal("finish")
		if lua_finish, ok := result.(*lua.LFunction); ok {
			if err := interpreter.CallByParam(lua.P{
				Fn:      lua_finish, // name of Lua function
				NRet:    0,          // number of returned values
				Protect: true,       // return err or panic
			}); err != nil {
				log.Fatal(err)
			}
		}

		interpreter.Close()

	}()

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
