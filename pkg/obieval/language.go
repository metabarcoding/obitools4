package obieval

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/goutils"
	"github.com/PaesslerAG/gval"
)

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
