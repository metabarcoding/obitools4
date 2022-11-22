package obieval

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/goutils"
	"github.com/PaesslerAG/gval"
)

func maxIntVector(values []int) float64 {
	m := values[0]
	for _,v := range values {
		if v > m {
			m = v
		}
	}

	return float64(m)
}

func maxIntMap(values  map[string]int) float64 {
	var m int
	first := true
	for _,v := range values {
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
	for _,v := range values {
		if v < m {
			m = v
		}
	}

	return float64(m)
}

func minIntMap(values  map[string]int) float64 {
	var m int
	first := true
	for _,v := range values {
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
	for _,v := range values {
		if v > m {
			m = v
		}
	}

	return m
}

func maxFloatMap(values  map[string]float64) float64 {
	var m float64
	first := true
	for _,v := range values {
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
	for _,v := range values {
		if v < m {
			m = v
		}
	}

	return m
}

func minFloatMap(values  map[string]float64) float64 {
	var m float64
	first := true
	for _,v := range values {
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
	}))

func Expression(expression string) (gval.Evaluable, error) {
	return OBILang.NewEvaluable(expression)
}
