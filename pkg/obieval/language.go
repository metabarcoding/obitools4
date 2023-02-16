package obieval

import (
	"fmt"
	"log"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/goutils"
	"github.com/PaesslerAG/gval"
)

func maxIntVector(values []int) float64 {
	m := values[0]
	for _, v := range values {
		if v > m {
			m = v
		}
	}

	return float64(m)
}

func maxIntMap(values map[string]int) float64 {
	var m int
	first := true
	for _, v := range values {
		if first {
			first = false
			m = v
		} else {
			if v > m {
				m = v
			}
		}
	}

	return float64(m)
}

func minIntVector(values []int) float64 {
	m := values[0]
	for _, v := range values {
		if v < m {
			m = v
		}
	}

	return float64(m)
}

func minIntMap(values map[string]int) float64 {
	var m int
	first := true
	for _, v := range values {
		if first {
			first = false
			m = v
		} else {
			if v < m {
				m = v
			}
		}
	}

	return float64(m)
}

func maxFloatVector(values []float64) float64 {
	m := values[0]
	for _, v := range values {
		if v > m {
			m = v
		}
	}

	return m
}

func maxFloatMap(values map[string]float64) float64 {
	var m float64
	first := true
	for _, v := range values {
		if first {
			first = false
			m = v
		} else {
			if v > m {
				m = v
			}
		}
	}

	return m
}

func minFloatVector(values []float64) float64 {
	m := values[0]
	for _, v := range values {
		if v < m {
			m = v
		}
	}

	return m
}

func minFloatMap(values map[string]float64) float64 {
	var m float64
	first := true
	for _, v := range values {
		if first {
			first = false
			m = v
		} else {
			if v < m {
				m = v
			}
		}
	}

	return m
}

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
		length := goutils.Len(args[0])
		return (float64)(length), nil
	}),
	gval.Function("ismap", func(args ...interface{}) (interface{}, error) {
		ismap := goutils.IsAMap(args[0])
		return ismap, nil
	}),
	gval.Function("printf", func(args ...interface{}) (interface{}, error) {
		text := fmt.Sprintf(args[0].(string), args[1:]...)
		return text, nil
	}),
	gval.Function("int", func(args ...interface{}) (interface{}, error) {
		val, err := goutils.InterfaceToInt(args[0])

		if err != nil {
			log.Fatalf("%v cannot be converted to an integer value", args[0])
		}
		return val, nil
	}),
	gval.Function("numeric", func(args ...interface{}) (interface{}, error) {
		val, err := goutils.InterfaceToFloat64(args[0])

		if err != nil {
			log.Fatalf("%v cannot be converted to a numeric value", args[0])
		}
		return val, nil
	}),
	gval.Function("bool", func(args ...interface{}) (interface{}, error) {
		val, err := goutils.InterfaceToBool(args[0])

		if err != nil {
			log.Fatalf("%v cannot be converted to a boolan value", args[0])
		}
		return val, nil
	}))

func Expression(expression string) (gval.Evaluable, error) {
	return OBILang.NewEvaluable(expression)
}
