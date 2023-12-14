package ncbitaxdump

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
)

func loadNodeTable(reader io.Reader, taxonomy *obitax.Taxonomy) {
	file := csv.NewReader(reader)
	file.Comma = '|'
	file.Comment = '#'
	file.TrimLeadingSpace = true
	file.ReuseRecord = true

	n := 0

	for record, err := file.Read(); err == nil; record, err = file.Read() {
		n++
		taxid, err := strconv.Atoi(strings.TrimSpace(record[0]))

		if err != nil {
			log.Panicf("Cannot read taxid at line %d: %v", n, err)
		}

		parent, err := strconv.Atoi(strings.TrimSpace(record[1]))

		if err != nil {
			log.Panicf("Cannot read parent taxid at line %d: %v", n, err)
		}

		rank := strings.TrimSpace(record[2])

		taxonomy.AddNewTaxa(taxid, parent, rank, true, true)
	}

	taxonomy.ReindexParent()
}

func loadNameTable(reader io.Reader, taxonomy *obitax.Taxonomy, onlysn bool) int {
	// file := csv.NewReader(reader)
	// file.Comma = '|'
	// file.Comment = '#'
	// file.TrimLeadingSpace = true
	// file.ReuseRecord = true
	// file.LazyQuotes = true
	file := bufio.NewReader(reader)

	n := 0

	for line, prefix, err := file.ReadLine(); err == nil; line, prefix, err = file.ReadLine() {

		if prefix {
			return -1
		}

		record := strings.Split(string(line), "|")
		taxid, _ := strconv.Atoi(strings.TrimSpace(record[0]))
		name := strings.TrimSpace(record[1])
		classname := strings.TrimSpace(record[3])

		if !onlysn || classname == "scientific name" {
			n++
			taxonomy.AddNewName(taxid, &name, &classname)
		}
	}

	return n
}

func loadMergedTable(reader io.Reader, taxonomy *obitax.Taxonomy) int {
	file := csv.NewReader(reader)
	file.Comma = '|'
	file.Comment = '#'
	file.TrimLeadingSpace = true
	file.ReuseRecord = true

	n := 0

	for record, err := file.Read(); err == nil; record, err = file.Read() {
		oldtaxid, _ := strconv.Atoi(strings.TrimSpace(record[0]))
		newtaxid, _ := strconv.Atoi(strings.TrimSpace(record[1]))
		n++
		taxonomy.AddNewAlias(newtaxid, oldtaxid)
	}

	return n
}

func LoadNCBITaxDump(directory string, onlysn bool) (*obitax.Taxonomy, error) {

	taxonomy := obitax.NewTaxonomy()

	//
	// Load the Taxonomy nodes
	//

	log.Printf("Loading Taxonomy nodes\n")

	nodefile, err := os.Open(path.Join(directory, "nodes.dmp"))
	if err != nil {
		return nil, fmt.Errorf("cannot open nodes file from '%s'",
			directory)
	}
	defer nodefile.Close()

	buffered := bufio.NewReader(nodefile)
	loadNodeTable(buffered, taxonomy)
	log.Printf("%d Taxonomy nodes read\n", taxonomy.Len())

	//
	// Load the Taxonomy nodes
	//

	log.Printf("Loading Taxon names\n")

	namefile, nerr := os.Open(path.Join(directory, "names.dmp"))
	if nerr != nil {
		return nil, fmt.Errorf("cannot open names file from '%s'",
			directory)
	}
	defer namefile.Close()

	n := loadNameTable(namefile, taxonomy, onlysn)
	log.Printf("%d taxon names read\n", n)

	//
	// Load the merged taxa
	//

	log.Printf("Loading Merged taxa\n")

	aliasfile, aerr := os.Open(path.Join(directory, "merged.dmp"))
	if aerr != nil {
		return nil, fmt.Errorf("cannot open merged file from '%s'",
			directory)
	}
	defer aliasfile.Close()

	buffered = bufio.NewReader(aliasfile)
	n = loadMergedTable(buffered, taxonomy)
	log.Printf("%d merged taxa read\n", n)

	return taxonomy, nil
}
