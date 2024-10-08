package obilua

import (
	log "github.com/sirupsen/logrus"

	lua "github.com/yuin/gopher-lua"
)

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
			v := table.RawGetInt(i)
			switch v.Type() {
			case lua.LTNil:
				val[i-1] = nil
			case lua.LTBool:
				val[i-1] = bool(v.(lua.LBool))
			case lua.LTNumber:
				val[i-1] = float64(v.(lua.LNumber))
			case lua.LTString:
				val[i-1] = string(v.(lua.LString))
			}
		}
		return val
	} else {
		// The table contains a hash
		val := make(map[string]interface{})
		table.ForEach(func(k, v lua.LValue) {
			if ks, ok := k.(lua.LString); ok {
				switch v.Type() {
				case lua.LTNil:
					val[string(ks)] = nil
				case lua.LTBool:
					val[string(ks)] = bool(v.(lua.LBool))
				case lua.LTNumber:
					val[string(ks)] = float64(v.(lua.LNumber))
				case lua.LTString:
					val[string(ks)] = string(v.(lua.LString))
				}
			}
		})
		return val
	}
}

func Table2ByteSlice(interpreter *lua.LState, table *lua.LTable) []byte {
	isArray := true
	table.ForEach(func(key, value lua.LValue) {
		if _, ok := key.(lua.LNumber); !ok {
			isArray = false
		}
	})

	if isArray {
		val := make([]byte, table.Len())

		for i := 1; i <= table.Len(); i++ {
			v := table.RawGetInt(i)
			switch v.Type() {
			case lua.LTNumber:
				if x := float64(v.(lua.LNumber)); x <= 255 {
					val[i-1] = byte(x)
				} else {
					log.Fatalf("LUA: Value %f at index %d is to large to be converted to byte", x, i)
				}
			default:
				log.Fatalf("LUA: Value %v at index %d cannot be converted to byte", v, i)
			}
		}

		return val
	}

	log.Fatalf("LUA: Value %v cannot be converted to []byte", table)

	return nil
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
