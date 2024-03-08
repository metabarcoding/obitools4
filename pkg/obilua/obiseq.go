package obilua

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	lua "github.com/yuin/gopher-lua"
)

func RegisterObiSeq(luaState *lua.LState) {
	registerBioSequenceType(luaState)
}

const luaBioSequenceTypeName = "BioSequence"

func registerBioSequenceType(luaState *lua.LState) {
	bioSequenceType := luaState.NewTypeMetatable(luaBioSequenceTypeName)
	luaState.SetGlobal(luaBioSequenceTypeName, bioSequenceType)
	luaState.SetField(bioSequenceType, "new", luaState.NewFunction(newObiSeq))

	luaState.SetField(bioSequenceType, "__index",
		luaState.SetFuncs(luaState.NewTable(),
			bioSequenceMethods))
}

func obiseq2Lua(interpreter *lua.LState,
	sequence *obiseq.BioSequence) lua.LValue {
	ud := interpreter.NewUserData()
	ud.Value = sequence
	interpreter.SetMetatable(ud, interpreter.GetTypeMetatable(luaBioSequenceTypeName))

	return ud
}

func newObiSeq(luaState *lua.LState) int {
	seqid := luaState.CheckString(1)
	seq := []byte(luaState.CheckString(2))

	definition := ""
	if luaState.GetTop() > 2 {
		definition = luaState.CheckString(3)
	}

	sequence := obiseq.NewBioSequence(seqid, seq, definition)

	luaState.Push(obiseq2Lua(luaState, sequence))
	return 1
}

var bioSequenceMethods = map[string]lua.LGFunction{
	"id":            bioSequenceGetSetId,
	"sequence":      bioSequenceGetSetSequence,
	"definition":    bioSequenceGetSetDefinition,
	"count":         bioSequenceGetSetCount,
	"taxid":         bioSequenceGetSetTaxid,
	"attribute":     bioSequenceGetSetAttribute,
	"len":           bioSequenceGetLength,
	"has_sequence":  bioSequenceHasSequence,
	"has_qualities": bioSequenceHasQualities,
	"source":        bioSequenceGetSource,
	"md5":           bioSequenceGetMD5,
}

// checkBioSequence checks if the first argument in the Lua stack is a *obiseq.BioSequence.
//
// This function accepts a pointer to the Lua state and attempts to retrieve a userdata
// that holds a pointer to a BioSequence. If the conversion is successful, it returns
// the *BioSequence. If the conversion fails, it raises a Lua argument error.
// Returns a pointer to obiseq.BioSequence or nil if the argument is not of the expected type.
func checkBioSequence(L *lua.LState) *obiseq.BioSequence {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*obiseq.BioSequence); ok {
		return v
	}
	L.ArgError(1, "obiseq.BioSequence expected")
	return nil
}

// bioSequenceGetSetId gets the ID of a biosequence or sets a new ID if provided.
//
// This function expects a *lua.LState pointer as its only parameter.
// If a second argument is provided, it sets the new ID for the biosequence.
// It returns 0 if a new ID is set, or 1 after pushing the current ID onto the stack.
func bioSequenceGetSetId(luaState *lua.LState) int {
	s := checkBioSequence(luaState)
	if luaState.GetTop() == 2 {
		s.SetId(luaState.CheckString(2))
		return 0
	}
	luaState.Push(lua.LString(s.Id()))
	return 1
}

func bioSequenceGetSetSequence(luaState *lua.LState) int {
	s := checkBioSequence(luaState)
	if luaState.GetTop() == 2 {
		s.SetSequence([]byte(luaState.CheckString(2)))
		return 0
	}
	luaState.Push(lua.LString(s.String()))
	return 1
}

func bioSequenceGetSetDefinition(luaState *lua.LState) int {
	s := checkBioSequence(luaState)
	if luaState.GetTop() == 2 {
		s.SetDefinition(luaState.CheckString(2))
		return 0
	}
	luaState.Push(lua.LString(s.Definition()))
	return 1
}

func bioSequenceGetSetCount(luaState *lua.LState) int {
	s := checkBioSequence(luaState)
	if luaState.GetTop() == 2 {
		s.SetCount(luaState.CheckInt(2))
		return 0
	}
	luaState.Push(lua.LNumber(s.Count()))
	return 1
}

func bioSequenceGetSetTaxid(luaState *lua.LState) int {
	s := checkBioSequence(luaState)
	if luaState.GetTop() == 2 {
		s.SetTaxid(luaState.CheckInt(2))
		return 0
	}
	luaState.Push(lua.LNumber(s.Taxid()))
	return 1
}

func bioSequenceGetSetAttribute(luaState *lua.LState) int {
	s := checkBioSequence(luaState)
	attName := luaState.CheckString(2)

	if luaState.GetTop() == 3 {
		ud := luaState.CheckAny(3)

		//
		// Perhaps the code needs some type checking on ud.Value
		// It's a first test
		//

		if v, ok := ud.(*lua.LTable); ok {
			s.SetAttribute(attName, Table2Interface(luaState, v))
		} else {
			s.SetAttribute(attName, ud)
		}

		return 0
	}

	value, ok := s.GetAttribute(attName)

	if !ok {
		luaState.Push(lua.LNil)
	} else {
		pushInterfaceToLua(luaState, value)
	}

	return 1
}

func bioSequenceGetLength(luaState *lua.LState) int {
	s := checkBioSequence(luaState)
	luaState.Push(lua.LNumber(s.Len()))
	return 1
}

func bioSequenceHasSequence(luaState *lua.LState) int {
	s := checkBioSequence(luaState)
	luaState.Push(lua.LBool(s.HasSequence()))
	return 1
}

func bioSequenceHasQualities(luaState *lua.LState) int {
	s := checkBioSequence(luaState)
	luaState.Push(lua.LBool(s.HasQualities()))
	return 1
}

func bioSequenceGetSource(luaState *lua.LState) int {
	s := checkBioSequence(luaState)
	if s.HasSource() {
		luaState.Push(lua.LString(s.Source()))
	} else {
		luaState.Push(lua.LNil)
	}
	return 1
}

func bioSequenceGetMD5(luaState *lua.LState) int {
	s := checkBioSequence(luaState)
	md5 := s.MD5()
	rt := luaState.NewTable()
	for i := 0; i < 16; i++ {
		rt.Append(lua.LNumber(md5[i]))
	}
	luaState.Push(rt)
	return 1
}
