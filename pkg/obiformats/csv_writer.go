package obiformats

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

func CSVSequenceRecord(sequence *obiseq.BioSequence, opt Options) []string {
	keys := opt.CSVKeys()
	record := make([]string, 0, len(keys)+4)

	if opt.CSVId() {
		record = append(record, sequence.Id())
	}

	if opt.CSVCount() {
		record = append(record, fmt.Sprint(sequence.Count()))
	}

	if opt.CSVTaxon() {
		taxid := sequence.Taxid()
		sn, ok := sequence.GetStringAttribute("scientific_name")

		if !ok {
			sn = opt.CSVNAValue()
		}

		record = append(record, fmt.Sprint(taxid), fmt.Sprint(sn))
	}

	if opt.CSVDefinition() {
		record = append(record, sequence.Definition())
	}

	for _, key := range opt.CSVKeys() {
		value, ok := sequence.GetAttribute(key)
		if !ok {
			value = opt.CSVNAValue()
		}

		svalue, _ := obiutils.InterfaceToString(value)
		record = append(record, svalue)
	}

	if opt.CSVSequence() {
		record = append(record, string(sequence.Sequence()))
	}

	if opt.CSVQuality() {
		if sequence.HasQualities() {
			l := sequence.Len()
			q := sequence.Qualities()
			ascii := make([]byte, l)
			quality_shift := obidefault.WriteQualitiesShift()
			for j := 0; j < l; j++ {
				ascii[j] = uint8(q[j]) + uint8(quality_shift)
			}
			record = append(record, string(ascii))
		} else {
			record = append(record, opt.CSVNAValue())
		}
	}

	return record
}
