package ncbitaxdump

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitax"
)

func loadNodeTable(reader io.Reader, taxonomy *obitax.Taxonomy) {
	file := csv.NewReader(reader)
	file.Comma = '|'
	file.Comment = '#'
	file.TrimLeadingSpace = true
	file.ReuseRecord = true

	for record, err := file.Read(); err == nil; record, err = file.Read() {
		taxid, _ := strconv.Atoi(strings.TrimSpace(record[0]))
		parent, _ := strconv.Atoi(strings.TrimSpace(record[1]))
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
		return nil, errors.New(fmt.Sprintf("Cannot open nodes file from '%s'",
			directory))
	}
	defer nodefile.Close()

	buffered := bufio.NewReader(nodefile)
	loadNodeTable(buffered, taxonomy)
	log.Printf("%d Taxonomy nodes read\n", taxonomy.Length())

	//
	// Load the Taxonomy nodes
	//

	log.Printf("Loading Taxon names\n")

	namefile, nerr := os.Open(path.Join(directory, "names.dmp"))
	if nerr != nil {
		return nil, errors.New(fmt.Sprintf("Cannot open names file from '%s'",
			directory))
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
		return nil, errors.New(fmt.Sprintf("Cannot open merged file from '%s'",
			directory))
	}
	defer aliasfile.Close()

	buffered = bufio.NewReader(aliasfile)
	n = loadMergedTable(buffered, taxonomy)
	log.Printf("%d merged taxa read\n", n)

	return taxonomy, nil
}
