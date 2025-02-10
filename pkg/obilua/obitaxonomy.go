package obilua

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	lua "github.com/yuin/gopher-lua"
)

func RegisterObiTaxonomy(luaState *lua.LState) {
	registerTaxonomyType(luaState)
	registerTaxonType(luaState)
}

const luaTaxonomyTypeName = "Taxonomy"

func registerTaxonomyType(luaState *lua.LState) {
	taxonomyType := luaState.NewTypeMetatable(luaTaxonomyTypeName)
	luaState.SetGlobal(luaTaxonomyTypeName, taxonomyType)
	luaState.SetField(taxonomyType, "new", luaState.NewFunction(newTaxonomy))
	luaState.SetField(taxonomyType, "default", luaState.NewFunction(defaultTaxonomy))
	luaState.SetField(taxonomyType, "has_default", luaState.NewFunction(hasDefaultTaxonomy))
	luaState.SetField(taxonomyType, "nil", taxon2Lua(luaState, nil))
	luaState.SetField(taxonomyType, "__index",
		luaState.SetFuncs(luaState.NewTable(),
			taxonomyMethods))
}

func taxonomy2Lua(interpreter *lua.LState,
	taxonomy *obitax.Taxonomy) lua.LValue {
	ud := interpreter.NewUserData()
	ud.Value = taxonomy
	interpreter.SetMetatable(ud, interpreter.GetTypeMetatable(luaTaxonomyTypeName))

	return ud
}

func newTaxonomy(luaState *lua.LState) int {
	name := luaState.CheckString(1)
	code := luaState.CheckString(2)

	charset := obiutils.AsciiAlphaNumSet
	if luaState.GetTop() > 2 {
		charset = obiutils.AsciiSetFromString(luaState.CheckString(3))
	}

	taxonomy := obitax.NewTaxonomy(name, code, charset)

	luaState.Push(taxonomy2Lua(luaState, taxonomy))
	return 1
}

func defaultTaxonomy(luaState *lua.LState) int {
	taxonomy := obitax.DefaultTaxonomy()

	if taxonomy == nil {
		luaState.RaiseError("No default taxonomy")
		return 0
	}

	luaState.Push(taxonomy2Lua(luaState, taxonomy))
	return 1
}

func hasDefaultTaxonomy(luaState *lua.LState) int {
	taxonomy := obitax.DefaultTaxonomy()

	luaState.Push(lua.LBool(taxonomy != nil))
	return 1
}

var taxonomyMethods = map[string]lua.LGFunction{
	"name":  taxonomyGetName,
	"code":  taxonomyGetCode,
	"taxon": taxonomyGetTaxon,
}

func checkTaxonomy(L *lua.LState) *obitax.Taxonomy {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*obitax.Taxonomy); ok {
		return v
	}
	L.ArgError(1, "obitax.Taxonomy expected")
	return nil
}

func taxonomyGetName(luaState *lua.LState) int {
	taxo := checkTaxonomy(luaState)
	luaState.Push(lua.LString(taxo.Name()))
	return 1
}

func taxonomyGetCode(luaState *lua.LState) int {
	taxo := checkTaxonomy(luaState)
	luaState.Push(lua.LString(taxo.Code()))
	return 1
}

func taxonomyGetTaxon(luaState *lua.LState) int {
	taxo := checkTaxonomy(luaState)
	taxid := luaState.CheckString(2)
	taxon, isAlias, err := taxo.Taxon(taxid)

	if err != nil {
		luaState.RaiseError("%s : Error on taxon taxon: %v", taxid, err)
		return 0
	}

	if isAlias && obidefault.FailOnTaxonomy() {
		luaState.RaiseError("%s : Taxon is an alias of %s", taxid, taxon.String())
		return 0
	}

	luaState.Push(taxon2Lua(luaState, taxon))
	return 1
}
