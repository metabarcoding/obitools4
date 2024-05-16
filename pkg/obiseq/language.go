package obiseq

import (
	"fmt"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	"github.com/PaesslerAG/gval"
)

// func maxIntVector(values []int) float64 {
// 	m := values[0]
// 	for _, v := range values {
// 		if v > m {
// 			m = v
// 		}
// 	}

// 	return float64(m)
// }

// func maxIntMap(values map[string]int) float64 {
// 	var m int
// 	first := true
// 	for _, v := range values {
// 		if first {
// 			first = false
// 			m = v
// 		} else {
// 			if v > m {
// 				m = v
// 			}
// 		}
// 	}

// 	return float64(m)
// }

// 	return float64(m)
// }

// func minIntMap(values map[string]int) float64 {
// 	var m int
// 	first := true
// 	for _, v := range values {
// 		if first {
// 			first = false
// 			m = v
// 		} else {
// 			if v < m {
// 				m = v
// 			}
// 		}
// 	}

// 	return float64(m)
// }

// func maxFloatVector(values []float64) float64 {
// 	m := values[0]
// 	for _, v := range values {
// 		if v > m {
// 			m = v
// 		}
// 	}

// 	return m
// }

// func maxFloatMap(values map[string]float64) float64 {
// 	var m float64
// 	first := true
// 	for _, v := range values {
// 		if first {
// 			first = false
// 			m = v
// 		} else {
// 			if v > m {
// 				m = v
// 			}
// 		}
// 	}

// 	return m
// }

// func minFloatVector(values []float64) float64 {
// 	m := values[0]
// 	for _, v := range values {
// 		if v < m {
// 			m = v
// 		}
// 	}

// 	return m
// }

// func minFloatMap(values map[string]float64) float64 {
// 	var m float64
// 	first := true
// 	for _, v := range values {
// 		if first {
// 			first = false
// 			m = v
// 		} else {
// 			if v < m {
// 				m = v
// 			}
// 		}
// 	}

// 	return m
// }

// func maxNumeric(args ...interface{}) (interface{}, error) {
// 	var m float64
//     first := true

// 	for _, v := range args {
// 		switch {
// 			case
// 		}
// 	}

// }

var OBILang = gval.NewLanguage(
	gval.Full(),
	gval.Function("len", func(args ...interface{}) (interface{}, error) {
		length := obiutils.Len(args[0])
		return (float64)(length), nil
	}),
	gval.Function("contains", func(args ...interface{}) (interface{}, error) {
		if obiutils.IsAMap(args[0]) {
			val := reflect.ValueOf(args[0]).MapIndex(reflect.ValueOf(args[1]))
			return val.IsValid(), nil
		}
		return false, nil
	}),
	gval.Function("ismap", func(args ...interface{}) (interface{}, error) {
		ismap := obiutils.IsAMap(args[0])
		return ismap, nil
	}),
	gval.Function("printf", func(args ...interface{}) (interface{}, error) {
		text := fmt.Sprintf(args[0].(string), args[1:]...)
		return text, nil
	}),
	gval.Function("gsub", func(args ...interface{}) (interface{}, error) {
		text := strings.ReplaceAll(args[0].(string), args[1].(string), args[2].(string))
		return text, nil
	}),
	gval.Function("subspc", func(args ...interface{}) (interface{}, error) {
		text := strings.ReplaceAll(args[0].(string), " ", "_")
		return text, nil
	}),
	gval.Function("int", func(args ...interface{}) (interface{}, error) {
		val, err := obiutils.InterfaceToInt(args[0])

		if err != nil {
			log.Fatalf("%v cannot be converted to an integer value", args[0])
		}
		return val, nil
	}),
	gval.Function("numeric", func(args ...interface{}) (interface{}, error) {
		val, err := obiutils.InterfaceToFloat64(args[0])

		if err != nil {
			log.Fatalf("%v cannot be converted to a numeric value", args[0])
		}
		return val, nil
	}),
	gval.Function("bool", func(args ...interface{}) (interface{}, error) {
		val, err := obiutils.InterfaceToBool(args[0])

		if err != nil {
			log.Fatalf("%v cannot be converted to a boolan value", args[0])
		}
		return val, nil
	}),
	gval.Function("ifelse", func(args ...interface{}) (interface{}, error) {
		if args[0].(bool) {
			return args[1], nil
		} else {
			return args[2], nil
		}
	}),
	gval.Function("gcskew", func(args ...interface{}) (interface{}, error) {
		composition := (args[0].(*BioSequence)).Composition()
		return float64(composition['g']-composition['c']) / float64(composition['g']+composition['c']), nil
	}),
	gval.Function("gc", func(args ...interface{}) (interface{}, error) {
		composition := (args[0].(*BioSequence)).Composition()
		return float64(composition['g']+composition['c']) / float64(args[0].(*BioSequence).Len()), nil
	}),
	gval.Function("composition", func(args ...interface{}) (interface{}, error) {
		comp := (args[0].(*BioSequence)).Composition()
		scomp := make(map[string]float64)
		for k, v := range comp {
			scomp[string(k)] = float64(v)
		}
		return scomp, nil
	}))
