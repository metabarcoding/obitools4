package obilua

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	lua "github.com/yuin/gopher-lua"
)

const luaTaxonTypeName = "Taxon"

func registerTaxonType(luaState *lua.LState) {
	taxonType := luaState.NewTypeMetatable(luaTaxonTypeName)
	luaState.SetGlobal(luaTaxonTypeName, taxonType)
	luaState.SetField(taxonType, "new", luaState.NewFunction(newTaxon))
	luaState.SetField(taxonType, "nil", taxonomy2Lua(luaState, nil))

	luaState.SetField(taxonType, "__index",
		luaState.SetFuncs(luaState.NewTable(),
			taxonMethods))
}

func taxon2Lua(interpreter *lua.LState,
	taxon *obitax.Taxon) lua.LValue {
	ud := interpreter.NewUserData()
	ud.Value = taxon
	interpreter.SetMetatable(ud, interpreter.GetTypeMetatable(luaTaxonTypeName))

	return ud
}

func newTaxon(luaState *lua.LState) int {
	taxonomy := checkTaxonomy(luaState)
	taxid := luaState.CheckString(2)
	parent := luaState.CheckString(3)
	sname := luaState.CheckString(4)
	rank := luaState.CheckString(5)

	isroot := false

	if luaState.GetTop() > 5 {
		isroot = luaState.CheckBool(6)
	}

	taxon, err := taxonomy.AddTaxon(taxid, parent, rank, isroot, false)

	if err != nil {
		luaState.RaiseError("(%v,%v,%v) : Error on taxon creation: %v", taxid, parent, sname, err)
		return 0
	}

	taxon.SetName(sname, "scientific name")

	luaState.Push(taxon2Lua(luaState, taxon))
	return 1
}

var taxonMethods = map[string]lua.LGFunction{
	"string":          taxonAsString,
	"scientific_name": taxonGetSetScientificName,
	"parent":          taxonGetParent,
	"taxon_at_rank":   taxGetTaxonAtRank,
	"species":         taxonGetSpecies,
	"genus":           taxonGetGenus,
	"family":          taxonGetFamily,
}

func checkTaxon(L *lua.LState, i int) *obitax.Taxon {
	ud := L.CheckUserData(i)
	if v, ok := ud.Value.(*obitax.Taxon); ok {
		return v
	}
	L.ArgError(i, "obitax.Taxon expected")
	return nil
}

func taxonAsString(luaState *lua.LState) int {
	taxon := checkTaxon(luaState, 1)
	luaState.Push(lua.LString(taxon.String()))
	return 1
}

func taxonGetSetScientificName(luaState *lua.LState) int {
	taxon := checkTaxon(luaState, 1)

	if luaState.GetTop() > 1 {
		sname := luaState.CheckString(2)
		taxon.SetName(sname, "scientific name")
		return 0
	}

	luaState.Push(lua.LString(taxon.ScientificName()))
	return 1
}

func taxonGetParent(luaState *lua.LState) int {
	taxon := checkTaxon(luaState, 1)

	parent := taxon.Parent()
	luaState.Push(taxon2Lua(luaState, parent))

	return 1
}

func taxonGetSpecies(luaState *lua.LState) int {
	taxon := checkTaxon(luaState, 1)

	species := taxon.Species()
	luaState.Push(taxon2Lua(luaState, species))

	return 1
}

func taxonGetGenus(luaState *lua.LState) int {
	taxon := checkTaxon(luaState, 1)

	genus := taxon.Genus()
	luaState.Push(taxon2Lua(luaState, genus))

	return 1
}

func taxonGetFamily(luaState *lua.LState) int {
	taxon := checkTaxon(luaState, 1)

	family := taxon.Family()
	luaState.Push(taxon2Lua(luaState, family))

	return 1
}

func taxGetTaxonAtRank(luaState *lua.LState) int {
	taxon := checkTaxon(luaState, 1)
	rank := luaState.CheckString(2)

	taxonAt := taxon.TaxonAtRank(rank)

	luaState.Push(taxon2Lua(luaState, taxonAt))

	return 1
}
