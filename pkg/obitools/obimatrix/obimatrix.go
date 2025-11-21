package obimatrix

import (
	"encoding/csv"
	"os"
	"slices"
	"sort"
	"sync"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obicsv"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	"golang.org/x/exp/maps"
)

type MatrixData struct {
	matrix        map[string]map[string]interface{}
	attributes    map[string]map[string]interface{}
	attributeList []string
	naValue       string
}

// MakeMatrixData generates a MatrixData instance.
//
// No parameters.
// Returns a MatrixData.
func MakeMatrixData(naValue string, attributes ...string) MatrixData {
	return MatrixData{
		matrix:        make(map[string]map[string]interface{}),
		attributes:    make(map[string]map[string]interface{}),
		attributeList: slices.Clone(attributes),
		naValue:       naValue,
	}
}

// NewMatrixData creates a new instance of MatrixData.
//
// It does not take any parameters.
// It returns a pointer to a MatrixData object.
func NewMatrixData(naValue string, attributes ...string) *MatrixData {
	m := MakeMatrixData(naValue, attributes...)
	return &m
}

// TransposeMatrixData transposes the MatrixData.
//
// It takes no parameters.
// If the input matrix has attributes, they are lost.
// A unique attribute "id" is added to store the column ids of the input matrix.
// It returns a pointer to the transposed MatrixData.
func (matrix *MatrixData) TransposeMatrixData() *MatrixData {
	m := MakeMatrixData(matrix.naValue, "id")
	for k, v := range matrix.matrix {
		for kk, vv := range v {
			log.Warnf("k = %s. kk=%s", k, kk)
			if _, ok := m.matrix[kk]; !ok {
				m.matrix[kk] = make(map[string]interface{})
			}
			m.matrix[kk][k] = vv
			m.attributes[kk] = map[string]interface{}{"id": kk}
		}
	}
	return &m
}

// MergeMatrixData merges the data from data2 into data1.
//
// data1 - Pointer to the MatrixData to merge into.
// data2 - Pointer to the MatrixData to merge from.
// Returns the pointer to the merged MatrixData.
func (data1 *MatrixData) MergeMatrixData(data2 *MatrixData) *MatrixData {

	for k := range data2.matrix {
		if _, ok := data1.matrix[k]; ok {
			log.Panicf("Sequence Id %s exists at least twice in the data set", k)
		} else {
			data1.matrix[k] = data2.matrix[k]
			data1.attributes[k] = data2.attributes[k]
		}
	}

	return data1
}

// Update updates the MatrixData with the given BioSequence and mapkey.
//
// Parameters:
// - s: The BioSequence object to update MatrixData with.
// - mapkey: The key to retrieve the attribute from the BioSequence object.
//
// Returns:
// - *MatrixData: The updated MatrixData object.
func (data *MatrixData) Update(s *obiseq.BioSequence, mapkey string, strict bool) *MatrixData {

	sid := s.Id()

	if _, ok := data.matrix[sid]; ok {
		log.Panicf("Sequence Id %s exists at least twice in the data set", sid)
	}
	if v, ok := s.GetAttribute(mapkey); ok {
		if m, ok := v.(*obiseq.StatsOnValues); ok {
			m.RLock()
			data.matrix[sid] = obiutils.MapToMapInterface(m.Map())
			m.RUnlock()
		} else if obiutils.IsAMap(v) {
			data.matrix[sid] = obiutils.MapToMapInterface(v)
		} else {
			log.Panicf("Attribute %s is not a map in the sequence %s", mapkey, s.Id())
		}
	} else {
		if strict {
			log.Panicf("Attribute %s does not exist in the sequence %s", mapkey, s.Id())
		}
		data.matrix[sid] = make(map[string]interface{})
	}

	attrs := make(map[string]interface{}, len(data.attributeList))
	for _, attrname := range data.attributeList {
		var value interface{}
		ok := false
		switch attrname {
		case "id":
			value = s.Id()
			ok = true
		case "count":
			value = s.Count()
			ok = true
		case "taxon":
			taxon := s.Taxon(nil)
			if taxon != nil {
				value = taxon.String()

			} else {
				value = s.Taxid()
			}
			ok = true
		case "sequence":
			value = s.String()
			ok = true
		case "quality":
			if s.HasQualities() {
				l := s.Len()
				q := s.Qualities()
				ascii := make([]byte, l)
				quality_shift := obidefault.WriteQualitiesShift()
				for j := 0; j < l; j++ {
					ascii[j] = uint8(q[j]) + uint8(quality_shift)
				}
				value = string(ascii)
				ok = true
			}
		default:
			value, ok = s.GetAttribute(attrname)
		}
		if ok {
			attrs[attrname] = value
		}
	}
	data.attributes[sid] = attrs

	return data
}

func IMatrix(iterator obiiter.IBioSequence) *MatrixData {

	nproc := obidefault.ParallelWorkers()
	waiter := sync.WaitGroup{}

	mapAttribute := CLIMapAttribute()
	attribList := make([]string, 0)

	if obicsv.CLIPrintId() {
		attribList = append(attribList, "id")
	}

	if obicsv.CLIPrintCount() {
		attribList = append(attribList, "count")
	}

	if obicsv.CLIPrintTaxon() {
		attribList = append(attribList, "taxon")
	}

	if obicsv.CLIPrintDefinition() {
		attribList = append(attribList, "definition")
	}

	if obicsv.CLIPrintSequence() {
		attribList = append(attribList, "sequence")
	}

	if obicsv.CLIPrintQuality() {
		attribList = append(attribList, "qualities")
	}

	attribList = append(attribList, obicsv.CLIToBeKeptAttributes()...)

	if obicsv.CLIAutoColumns() {
		if iterator.Next() {
			batch := iterator.Get()
			if len(batch.Slice()) == 0 {
				log.Panicf("first batch should not be empty")
			}
			auto_slot := batch.Slice().AttributeKeys(true, true).Members()
			slices.Sort(auto_slot)
			attribList = append(attribList, auto_slot...)
			iterator.PushBack()
		}
	}

	naValue := obicsv.CLINAValue()
	strict := CLIStrict()

	summaries := make([]*MatrixData, nproc)

	ff := func(iseq obiiter.IBioSequence, summary *MatrixData) {

		for iseq.Next() {
			batch := iseq.Get()
			for _, seq := range batch.Slice() {
				summary.Update(seq, mapAttribute, strict)
			}
		}
		waiter.Done()
	}

	waiter.Add(nproc)

	summaries[0] = NewMatrixData(naValue, attribList...)
	go ff(iterator, summaries[0])

	for i := 1; i < nproc; i++ {
		summaries[i] = NewMatrixData(naValue, attribList...)
		go ff(iterator.Split(), summaries[i])
	}

	waiter.Wait()
	obiutils.WaitForLastPipe()

	rep := summaries[0]

	for i := 1; i < nproc; i++ {
		rep = rep.MergeMatrixData(summaries[i])
	}

	return rep
}

func CLIWriteCSVToStdout(matrix *MatrixData) {
	navalue := CLIMapNaValue()
	csvwriter := csv.NewWriter(os.Stdout)

	if CLITranspose() {
		matrix = matrix.TransposeMatrixData()
	}

	samples := obiutils.NewSet[string]()

	for _, v := range matrix.matrix {
		samples.Add(maps.Keys(v)...)
	}

	osamples := samples.Members()
	sort.Strings(osamples)
	columns := make([]string, 0, len(osamples)+len(matrix.attributeList))
	columns = append(columns, matrix.attributeList...)
	columns = append(columns, osamples...)

	header := slices.Clone(columns)

	csvwriter.Write(columns)
	nattribs := len(matrix.attributeList)

	for k, data := range matrix.matrix {
		attrs := matrix.attributes[k]
		for i, kk := range header {
			if i < nattribs {
				if v, ok := attrs[kk]; ok {
					vs, err := obiutils.InterfaceToString(v)
					if err != nil {
						log.Panicf("value  %v in sequence %s for attribute %s cannot be casted to a string", v, k, kk)
					}
					columns[i] = vs
				} else {
					columns[i] = matrix.naValue
				}
			} else {
				if v, ok := data[kk]; ok {
					vs, err := obiutils.InterfaceToString(v)
					if err != nil {
						log.Panicf("value %v in sequence %s for attribute %s cannot be casted to a string", v, k, kk)
					}
					columns[i] = vs
				} else {
					columns[i] = navalue
				}
			}
		}
		csvwriter.Write(columns)
	}

	csvwriter.Flush()
}

func CLIWriteThreeColumnsToStdout(matrix *MatrixData) {
	sname := CLISampleName()
	vname := CLIValueName()
	csvwriter := csv.NewWriter(os.Stdout)

	csvwriter.Write([]string{"id", sname, vname})
	for seqid := range matrix.matrix {
		for attr, v := range matrix.matrix[seqid] {
			vs, err := obiutils.InterfaceToString(v)
			if err != nil {
				log.Panicf("value %v in sequence %s for attribute %s cannot be casted to a string", v, seqid, attr)
			}
			csvwriter.Write([]string{seqid, attr, vs})
		}
	}

	csvwriter.Flush()
}
