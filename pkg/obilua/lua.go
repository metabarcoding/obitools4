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

// NewInterpreter creates a new Lua interpreter and registers the Obilib and ObiContext modules.
//
// No parameters.
// Returns a pointer to a Lua state.
func NewInterpreter() *lua.LState {
	lua := lua.NewState()

	RegisterObilib(lua)
	RegisterObiContext(lua)

	return lua
}

// Compile compiles a Lua program into a Lua function proto.
//
// It takes a byte slice containing the Lua program and a string representing the name of the program.
// It returns a pointer to a Lua function proto and an error if any.
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

// CompileScript compiles a Lua script from a file.
//
// It takes a file path as input and returns a pointer to a Lua function proto and an error if any.
// The function reads the contents of the file specified by the file path and compiles it into a Lua function proto using the Compile function.
// If there is an error reading the file, the function returns nil and the error.
// Otherwise, it returns the compiled Lua function proto and nil error.
func CompileScript(filePath string) (*lua.FunctionProto, error) {
	program, err := os.ReadFile(filePath)

	if err != nil {
		return nil, err
	}

	return Compile(program, filePath)
}

// LuaWorker creates a Go function that executes a Lua script and returns a SeqWorker.
//
// The function takes a Lua function prototype as input and creates a new interpreter.
// It then creates a new Lua function from the prototype and pushes it onto the interpreter's stack.
// The interpreter calls the Lua function and checks for any errors.
// It retrieves the global variable "worker" from the interpreter and checks if it is a Lua function.
// If it is a Lua function, it defines a Go function that takes a BioSequence as input.
// Inside the Go function, it calls the Lua function with the BioSequence as an argument.
// It retrieves the result from the interpreter and checks its type.
// If the result is a BioSequence or a BioSequenceSlice, it returns it along with any error.
// If the result is not of the expected type, it returns an error.
// If the global variable "worker" is not a Lua function, it logs a fatal error.
// The Go function is returned as a SeqWorker.
//
// Parameters:
// - proto: The Lua function prototype.
//
// Return type:
// - obiseq.SeqWorker: The Go function that executes the Lua script and returns a SeqWorker.
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
				case *obiseq.BioSequenceSlice:
					return *val, err
				default:
					return nil, fmt.Errorf("worker function doesn't return the correct type")
				}
			}

			return nil, fmt.Errorf("worker function doesn't return the correct type")
		}

		return f
	}

	log.Fatalf("The worker object is not a function")
	return nil
}

// LuaProcessor processes a Lua script on a sequence iterator and returns a new iterator.
//
// Parameters:
// - iterator: The IBioSequence iterator to process.
// - name: The name of the Lua script.
// - program: The Lua script program as a string.
// - breakOnError: A boolean indicating whether to stop processing if an error occurs.
// - nworkers: An integer representing the number of workers to use for processing.
// Returns:
// - obiiter.IBioSequence: The new IBioSequence iterator after processing the Lua script.
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

// LuaPipe creates a pipeable function that applies a Lua script to an input sequence.
//
// Parameters:
// - name: The name of the Lua script.
// - program: The Lua script program as a string.
// - breakOnError: A boolean indicating whether to stop processing if an error occurs.
// - nworkers: An integer representing the number of workers to use for processing.
// Returns:
// - obiiter.Pipeable: A pipeable function that applies the Lua script to the input sequence.
func LuaPipe(name, program string, breakOnError bool, nworkers int) obiiter.Pipeable {

	f := func(input obiiter.IBioSequence) obiiter.IBioSequence {
		return LuaProcessor(input, name, program, breakOnError, nworkers)
	}

	return f
}

// LuaScriptPipe creates a pipeable function that applies a Lua script to an input sequence.
//
// Parameters:
// - filename: The name of the Lua script file.
// - breakOnError: A boolean indicating whether to stop processing if an error occurs.
// - nworkers: An integer representing the number of workers to use for processing.
// Returns:
// - obiiter.Pipeable: A pipeable function that applies the Lua script to the input sequence.
func LuaScriptPipe(filename string, breakOnError bool, nworkers int) obiiter.Pipeable {
	program, err := os.ReadFile(filename)

	if err != nil {
		log.Fatalf("Cannot read script file %s", filename)
	}

	return LuaPipe(filename, string(program), breakOnError, nworkers)
}
