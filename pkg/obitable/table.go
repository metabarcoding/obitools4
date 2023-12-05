// obitable provide a row oriented data table structure
package obitable

import (
	"reflect"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"github.com/chen3feng/stl4go"
)

type Header stl4go.Ordered

type Row map[string]interface{}
type Table struct {
	ColType map[string]reflect.Type
	Rows    []Row
}

type RowGetter func(name string) interface{}

func RowFromMap(data map[string]interface{}, navalue string) RowGetter {
	getter := func(name string) interface{} {
		value, ok := data[name]

		if !ok {
			value = navalue
		}

		return value
	}

	return getter
}

func RowFromBioSeq(data *obiseq.BioSequence, navalue string) RowGetter {
	getter := func(name string) interface{} {
		var value interface{}
		value = navalue

		switch name {
		case "id":
			value = data.Id()
		case "sequence":
			value = data.Sequence()
		case "definition":
			value = data.Definition()
		case "taxid":
			value = data.Taxid()
		case "count":
			value = data.Count()
		default:
			if data.HasAnnotation() {
				var ok bool
				value, ok = data.GetAttribute(name)
				if !ok {
					value = navalue
				}
			}
		}
		return value
	}

	return getter
}
