package obitax

import (
	"archive/tar"
	"bufio"
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"

	log "github.com/sirupsen/logrus"
)

func IsNCBITarTaxDump(path string) bool {

	file, err := obiutils.Ropen(path)

	if err != nil {
		return false
	}

	defer file.Close()

	citations := false
	division := false
	gencode := false
	names := false
	delnodes := false
	gc := false
	merged := false
	nodes := false

	tarfile := tar.NewReader(file)

	header, err := tarfile.Next()

	for err == nil {
		name := header.Name

		if header.Typeflag == tar.TypeReg {
			switch name {
			case "citations.dmp":
				citations = true
			case "division.dmp":
				division = true
			case "gencode.dmp":
				gencode = true
			case "names.dmp":
				names = true
			case "delnodes.dmp":
				delnodes = true
			case "gc.prt":
				gc = true
			case "merged.dmp":
				merged = true
			case "nodes.dmp":
				nodes = true
			}
		}
		header, err = tarfile.Next()
	}

	return citations && division && gencode && names && delnodes && gc && merged && nodes
}

func LoadNCBITarTaxDump(path string, onlysn bool) (*Taxonomy, error) {

	taxonomy := NewTaxonomy("NCBI Taxonomy", "taxon", obiutils.AsciiDigitSet)

	//
	// Load the Taxonomy nodes
	//

	log.Printf("Loading Taxonomy nodes\n")

	file, err := obiutils.Ropen(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open taxonomy file from '%s'",
			path)
	}

	nodefile, err := obiutils.TarFileReader(file, "nodes.dmp")
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("cannot open nodes file from '%s'",
			path)
	}

	buffered := bufio.NewReader(nodefile)
	loadNodeTable(buffered, taxonomy)
	log.Printf("%d Taxonomy nodes read\n", taxonomy.Len())
	file.Close()

	//
	// Load the Taxonomy nodes
	//

	log.Printf("Loading Taxon names\n")

	file, err = obiutils.Ropen(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open taxonomy file from '%s'",
			path)
	}

	namefile, nerr := obiutils.TarFileReader(file, "names.dmp")
	if nerr != nil {
		file.Close()
		return nil, fmt.Errorf("cannot open names file from '%s'",
			path)
	}
	n := loadNameTable(namefile, taxonomy, onlysn)
	log.Printf("%d taxon names read\n", n)
	file.Close()

	//
	// Load the merged taxa
	//

	log.Printf("Loading Merged taxa\n")
	file, err = obiutils.Ropen(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open taxonomy file from '%s'",
			path)
	}

	aliasfile, aerr := obiutils.TarFileReader(file, "merged.dmp")
	if aerr != nil {
		file.Close()
		return nil, fmt.Errorf("cannot open merged file from '%s'",
			path)
	}

	buffered = bufio.NewReader(aliasfile)
	n = loadMergedTable(buffered, taxonomy)
	log.Printf("%d merged taxa read\n", n)

	root, _, err := taxonomy.Taxon("1")

	if err != nil {
		log.Fatal("cannot find the root taxon (1) in the NCBI tax dump")
	}

	taxonomy.SetRoot(root)

	return taxonomy, nil
}
