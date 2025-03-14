package obiformats

import (
	"encoding/csv"
	"errors"
	"strings"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

func LoadCSVTaxonomy(path string, onlysn bool) (*obitax.Taxonomy, error) {

	log.Infof("Loading taxonomy from csv file: %s", path)

	file, err := obiutils.Ropen(path)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	csvfile := csv.NewReader(file)

	csvfile.Comma = ','
	csvfile.ReuseRecord = false
	csvfile.LazyQuotes = true
	csvfile.Comment = '#'
	csvfile.FieldsPerRecord = -1
	csvfile.TrimLeadingSpace = true

	header, err := csvfile.Read()

	if err != nil {
		log.Fatal(err)
	}

	taxidColIndex := -1
	parentColIndex := -1
	scientific_nameColIndex := -1
	rankColIndex := -1

	for i, colName := range header {
		switch colName {
		case "taxid":
			taxidColIndex = i
		case "parent":
			parentColIndex = i
		case "scientific_name":
			scientific_nameColIndex = i
		case "taxonomic_rank":
			rankColIndex = i
		}
	}

	if taxidColIndex == -1 {
		return nil, errors.New("taxonomy file does not contain taxid column")
	}

	if parentColIndex == -1 {
		return nil, errors.New("taxonomy file does not contain parent column")
	}

	if scientific_nameColIndex == -1 {
		return nil, errors.New("taxonomy file does not contain scientific_name column")
	}

	if rankColIndex == -1 {
		return nil, errors.New("taxonomy file does not contain rank column")
	}

	name := obiutils.RemoveAllExt(path)
	short := obiutils.Basename(path)

	line, err := csvfile.Read()
	if err == nil {
		parts := strings.Split(line[taxidColIndex], " ")
		parts = strings.Split(parts[0], ":")
		if len(parts) > 1 {
			short = parts[0]
		}
	}

	log.Infof("Taxonomy name: %s", name)
	log.Infof("Taxon code: %s", short)

	taxonomy := obitax.NewTaxonomy(name, short, obiutils.AsciiAlphaNumSet)

	root := true
	var taxon *obitax.Taxon

	for err == nil {
		taxid := line[taxidColIndex]
		parent := line[parentColIndex]
		scientific_name := line[scientific_nameColIndex]
		rank := line[rankColIndex]

		taxon, err = taxonomy.AddTaxon(taxid, parent, rank, root, false)

		if err != nil {
			log.Fatalf("cannot add taxon %s:  %v", taxid, err)
		}

		root = false

		taxon.SetName(scientific_name, "scientific name")

		line, err = csvfile.Read()
	}

	log.Infof("%d Taxa loaded", taxonomy.Len())

	if !taxonomy.HasRoot() {
		return nil, errors.New("taxonomy file does not contain root node")
	}

	return taxonomy, nil
}
