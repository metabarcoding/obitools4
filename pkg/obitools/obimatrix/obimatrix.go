package obimatrix

import (
	"encoding/csv"
	"os"
	"sort"
	"sync"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	"golang.org/x/exp/maps"
)

type MatrixData map[string]map[string]interface{}

// MakeMatrixData generates a MatrixData instance.
//
// No parameters.
// Returns a MatrixData.
func MakeMatrixData() MatrixData {
	return make(MatrixData)
}

// NewMatrixData creates a new instance of MatrixData.
//
// It does not take any parameters.
// It returns a pointer to a MatrixData object.
func NewMatrixData() *MatrixData {
	m := make(MatrixData)
	return &m
}

// TransposeMatrixData transposes the MatrixData.
//
// It takes no parameters.
// It returns a pointer to the transposed MatrixData.
func (matrix *MatrixData) TransposeMatrixData() *MatrixData {
	m := make(MatrixData)
	for k, v := range *matrix {
		for kk, vv := range v {
			if _, ok := m[kk]; !ok {
				m[kk] = make(map[string]interface{})
			}
			m[kk][k] = vv
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

	for k := range *data2 {
		if _, ok := (*data1)[k]; ok {
			log.Panicf("Sequence Id %s exists at least twice in the data set", k)
		} else {
			(*data1)[k] = (*data2)[k]
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
func (data *MatrixData) Update(s *obiseq.BioSequence, mapkey string) *MatrixData {
	if v, ok := s.GetAttribute(mapkey); ok {
		if m, ok := v.(*obiseq.StatsOnValues); ok {
			m.RLock()
			(*data)[s.Id()] = obiutils.MapToMapInterface(m.Map())
			m.RUnlock()
		} else if obiutils.IsAMap(v) {
			(*data)[s.Id()] = obiutils.MapToMapInterface(v)
		} else {
			log.Panicf("Attribute %s is not a map in the sequence %s", mapkey, s.Id())
		}
	} else {
		log.Panicf("Attribute %s does not exist in the sequence %s", mapkey, s.Id())
	}

	return data
}

func IMatrix(iterator obiiter.IBioSequence) *MatrixData {

	nproc := obidefault.ParallelWorkers()
	waiter := sync.WaitGroup{}

	mapAttribute := CLIMapAttribute()

	summaries := make([]*MatrixData, nproc)

	ff := func(iseq obiiter.IBioSequence, summary *MatrixData) {

		for iseq.Next() {
			batch := iseq.Get()
			for _, seq := range batch.Slice() {
				summary.Update(seq, mapAttribute)
			}
		}
		waiter.Done()
	}

	waiter.Add(nproc)

	summaries[0] = NewMatrixData()
	go ff(iterator, summaries[0])

	for i := 1; i < nproc; i++ {
		summaries[i] = NewMatrixData()
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
	navalue := CLINaValue()
	csvwriter := csv.NewWriter(os.Stdout)

	if CLITranspose() {
		matrix = matrix.TransposeMatrixData()
	}

	samples := obiutils.NewSet[string]()

	for _, v := range *matrix {
		samples.Add(maps.Keys(v)...)
	}

	osamples := samples.Members()
	sort.Strings(osamples)

	columns := make([]string, 1, len(osamples)+1)
	columns[0] = "id"
	columns = append(columns, osamples...)

	csvwriter.Write(columns)

	for k, data := range *matrix {
		columns = columns[0:1]
		columns[0] = k
		for _, kk := range osamples {
			if v, ok := data[kk]; ok {
				vs, err := obiutils.InterfaceToString(v)
				if err != nil {
					log.Panicf("value %v in sequence %s for attribute %s cannot be casted to a string", v, k, kk)
				}
				columns = append(columns, vs)
			} else {
				columns = append(columns, navalue)
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
	for seqid := range *matrix {
		for attr, v := range (*matrix)[seqid] {
			vs, err := obiutils.InterfaceToString(v)
			if err != nil {
				log.Panicf("value %v in sequence %s for attribute %s cannot be casted to a string", v, seqid, attr)
			}
			csvwriter.Write([]string{seqid, attr, vs})
		}
	}

	csvwriter.Flush()
}
