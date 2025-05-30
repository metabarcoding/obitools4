package obilua

import (
	"strings"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiformats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	lua "github.com/yuin/gopher-lua"
)

const luaBioSequenceSliceTypeName = "BioSequenceSlice"

func registerBioSequenceSliceType(luaState *lua.LState) {
	bioSequenceSliceType := luaState.NewTypeMetatable(luaBioSequenceSliceTypeName)
	luaState.SetGlobal(luaBioSequenceSliceTypeName, bioSequenceSliceType)
	luaState.SetField(bioSequenceSliceType, "new", luaState.NewFunction(newObiSeqSlice))
	luaState.SetField(bioSequenceSliceType, "nil", obiseqslice2Lua(luaState, nil))

	luaState.SetField(bioSequenceSliceType, "__index",
		luaState.SetFuncs(luaState.NewTable(),
			bioSequenceSliceMethods))
}

func obiseqslice2Lua(interpreter *lua.LState,
	seqslice *obiseq.BioSequenceSlice) lua.LValue {
	ud := interpreter.NewUserData()
	ud.Value = seqslice
	interpreter.SetMetatable(ud, interpreter.GetTypeMetatable(luaBioSequenceSliceTypeName))

	return ud
}

func newObiSeqSlice(luaState *lua.LState) int {
	seqslice := obiseq.NewBioSequenceSlice()
	luaState.Push(obiseqslice2Lua(luaState, seqslice))
	return 1
}

var bioSequenceSliceMethods = map[string]lua.LGFunction{
	"push":     bioSequenceSlicePush,
	"pop":      bioSequenceSlicePop,
	"sequence": bioSequenceSliceGetSetSequence,
	"len":      bioSequenceSliceGetLength,
	"fasta":    bioSequenceSliceGetFasta,
	"fastq":    bioSequenceSliceGetFastq,
	"string":   bioSequenceSliceAsString,
}

func checkBioSequenceSlice(L *lua.LState) *obiseq.BioSequenceSlice {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*obiseq.BioSequenceSlice); ok {
		return v
	}
	L.ArgError(1, "obiseq.BioSequenceSlice expected")
	return nil
}

func bioSequenceSliceGetLength(luaState *lua.LState) int {
	s := checkBioSequenceSlice(luaState)
	luaState.Push(lua.LNumber(s.Len()))
	return 1
}

func bioSequenceSliceGetSetSequence(luaState *lua.LState) int {
	s := checkBioSequenceSlice(luaState)
	index := luaState.CheckInt(2)

	if index > s.Len() || index < 0 {
		luaState.RaiseError("BioSequenceSlice index out of range")
		return 0
	}

	if luaState.GetTop() == 3 {
		ud := luaState.CheckUserData(3)
		if v, ok := ud.Value.(*obiseq.BioSequence); ok {
			(*s)[index] = v
			return 0
		}
		luaState.ArgError(1, "obiseq.BioSequenceSlice expected")
		return 0
	}

	value := obiseq2Lua(luaState, (*s)[index])
	luaState.Push(value)

	return 1
}

func bioSequenceSlicePush(luaState *lua.LState) int {
	s := checkBioSequenceSlice(luaState)

	ud := luaState.CheckUserData(2)
	if v, ok := ud.Value.(*obiseq.BioSequence); ok {
		(*s) = append((*s), v)
		return 0
	}

	luaState.ArgError(1, "obiseq.BioSequenceSlice expected")
	return 0
}

func bioSequenceSlicePop(luaState *lua.LState) int {
	s := checkBioSequenceSlice(luaState)
	if len(*s) == 0 {
		return 0
	}

	seq := (*s)[len(*s)-1]
	(*s) = (*s)[0 : len(*s)-1]
	value := obiseq2Lua(luaState, seq)
	luaState.Push(value)
	return 1

}

func bioSequenceSliceGetFasta(luaState *lua.LState) int {
	s := checkBioSequenceSlice(luaState)

	formater := obiformats.FormatFastSeqJsonHeader

	if luaState.GetTop() > 1 {
		format := luaState.CheckString(2)
		switch format {
		case "json":
			formater = obiformats.FormatFastSeqJsonHeader
		case "obi":
			formater = obiformats.FormatFastSeqOBIHeader
		}
	}

	txts := make([]string, len(*s))

	for i, seq := range *s {
		txts[i] = obiformats.FormatFasta(seq, formater)
	}

	txt := strings.Join(txts, "\n")

	luaState.Push(lua.LString(txt))
	return 1
}

func bioSequenceSliceGetFastq(luaState *lua.LState) int {
	s := checkBioSequenceSlice(luaState)

	formater := obiformats.FormatFastSeqJsonHeader

	if luaState.GetTop() > 1 {
		format := luaState.CheckString(2)
		switch format {
		case "json":
			formater = obiformats.FormatFastSeqJsonHeader
		case "obi":
			formater = obiformats.FormatFastSeqOBIHeader
		}
	}

	txts := make([]string, len(*s))

	for i, seq := range *s {
		txts[i] = obiformats.FormatFastq(seq, formater)
	}

	txt := strings.Join(txts, "\n")

	luaState.Push(lua.LString(txt))
	return 1
}

func bioSequenceSliceAsString(luaState *lua.LState) int {
	s := checkBioSequenceSlice(luaState)

	formater := obiformats.FormatFastSeqJsonHeader

	if luaState.GetTop() > 1 {
		format := luaState.CheckString(2)
		switch format {
		case "json":
			formater = obiformats.FormatFastSeqJsonHeader
		case "obi":
			formater = obiformats.FormatFastSeqOBIHeader
		}
	}

	txts := make([]string, len(*s))

	format := obiformats.FormatFasta

	allQual := true

	for _, s := range *s {
		allQual = allQual && s.HasQualities()
	}

	if allQual {
		format = obiformats.FormatFastq
	}

	for i, seq := range *s {
		txts[i] = format(seq, formater)
	}

	txt := strings.Join(txts, "\n")

	luaState.Push(lua.LString(txt))
	return 1
}
