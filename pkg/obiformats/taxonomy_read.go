package obiformats

import (
	"fmt"
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	"github.com/gabriel-vasile/mimetype"

	log "github.com/sirupsen/logrus"
)

type TaxonomyLoader func(path string, onlysn, seqAsTaxa bool) (*obitax.Taxonomy, error)

func DetectTaxonomyTarFormat(path string) (TaxonomyLoader, error) {

	switch {
	case IsNCBITarTaxDump(path):
		log.Infof("NCBI Taxdump Tar Archive detected: %s", path)
		return LoadNCBITarTaxDump, nil
	}

	return nil, fmt.Errorf("unknown taxonomy format: %s", path)
}

func DetectTaxonomyFormat(path string) (TaxonomyLoader, error) {

	obiutils.RegisterOBIMimeType()

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	file.Close()

	if fileInfo.IsDir() {
		// For the moment, we only support NCBI Taxdump directory format
		log.Infof("NCBI Taxdump detected: %s", path)
		return LoadNCBITaxDump, nil
	} else {
		file, err := obiutils.Ropen(path)

		if err != nil {
			return nil, err
		}

		mimetype, err := mimetype.DetectReader(file)

		if err != nil {
			file.Close()
			return nil, err
		}

		file.Close()

		switch mimetype.String() {
		case "text/csv":
			return LoadCSVTaxonomy, nil
		case "application/x-tar":
			return DetectTaxonomyTarFormat(path)
		case "text/fasta":
			return func(path string, onlysn, seqAsTaxa bool) (*obitax.Taxonomy, error) {
				input, err := ReadFastaFromFile(path)
				input = input.NumberSequences(1, true)

				if err != nil {
					return nil, err
				}
				_, data := input.Load()

				return data.ExtractTaxonomy(nil, seqAsTaxa)
			}, nil
		case "text/fastq":
			return func(path string, onlysn, seqAsTaxa bool) (*obitax.Taxonomy, error) {
				input, err := ReadFastqFromFile(path)
				input = input.NumberSequences(1, true)

				if err != nil {
					return nil, err
				}
				_, data := input.Load()

				return data.ExtractTaxonomy(nil, seqAsTaxa)
			}, nil
		}

		log.Fatalf("Detected file format: %s", mimetype.String())
	}

	return nil, nil
}

func LoadTaxonomy(path string, onlysn, seqAsTaxa bool) (*obitax.Taxonomy, error) {
	loader, err := DetectTaxonomyFormat(path)

	if err != nil {
		return nil, err
	}

	taxonomy, err := loader(path, onlysn, seqAsTaxa)

	return taxonomy, err
}
