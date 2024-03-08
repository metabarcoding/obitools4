package obilua

import lua "github.com/yuin/gopher-lua"

func Table2Interface(interpreter *lua.LState, table *lua.LTable) interface{} {
	// <EC> 07/03/2024: il y a sans doute plus efficace mais pour l'instant
	//                  ça marche
	isArray := true
	table.ForEach(func(key, value lua.LValue) {
		if _, ok := key.(lua.LNumber); !ok {
			isArray = false
		}
	})
	if isArray {
		val := make([]interface{}, table.Len())
		for i := 1; i <= table.Len(); i++ {
			val[i-1] = table.RawGetInt(i)
		}
		return val
	} else {
		// The table contains a hash
		val := make(map[string]interface{})
		table.ForEach(func(k, v lua.LValue) {
			if ks, ok := k.(lua.LString); ok {
				val[string(ks)] = v
			}
		})
		return val
	}
}

// 	}

// 	return nil
// }

// 	if x := table.RawGetInt(1); x != nil {
// 		val := make([]interface{}, table.Len())
// 		for i := 1; i <= table.Len(); i++ {
// 			val[i-1] = table.RawGetInt(i)
// 		}
// 		return val
// 	} else {

// 	}
// }

// if lv.Type() == lua.LTTable {
//     table := lv.(*lua.LTable)
//     isArray := true
//     table.ForEach(func(key, value lua.LValue) {
//         if _, ok := key.(lua.LNumber); !ok {
//             isArray = false
//         }
//     })
//     if isArray {
//         // The table contains an array
//     } else {
//         // The table contains a hash
//     }
// }
