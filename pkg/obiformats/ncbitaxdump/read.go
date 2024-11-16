package ncbitaxdump

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path"
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
		taxid := strings.TrimSpace(record[0])
		parent := strings.TrimSpace(record[1])
		rank := strings.TrimSpace(record[2])

		_, err := taxonomy.AddTaxon(taxid, parent, rank, taxid == "1", false)

		if err != nil {
			log.Fatalf("Error adding taxon %s: %v\n", taxid, err)
		}
	}
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
	l := 0

	for line, prefix, err := file.ReadLine(); err == nil; line, prefix, err = file.ReadLine() {
		l++
		if prefix {
			return -1
		}

		record := strings.Split(string(line), "|")
		taxid := strings.TrimSpace(record[0])

		name := strings.TrimSpace(record[1])
		classname := strings.TrimSpace(record[3])

		if !onlysn || classname == "scientific name" {
			n++
			taxonomy.Taxon(taxid).SetName(name, classname)
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
		n++
		oldtaxid := strings.TrimSpace(record[0])
		newtaxid := strings.TrimSpace(record[1])

		taxonomy.AddAlias(newtaxid, oldtaxid, false)
	}

	return n
}

// LoadNCBITaxDump loads the NCBI taxonomy data from the specified directory.
// It reads the taxonomy nodes, taxon names, and merged taxa from the corresponding files
// and constructs a Taxonomy object.
//
// Parameters:
//   - directory: A string representing the path to the directory containing the NCBI taxonomy dump files.
//   - onlysn: A boolean indicating whether to load only scientific names (true) or all names (false).
//
// Returns:
//   - A pointer to the obitax.Taxonomy object containing the loaded taxonomy data, or an error
//     if any of the files cannot be opened or read.
func LoadNCBITaxDump(directory string, onlysn bool) (*obitax.Taxonomy, error) {

	taxonomy := obitax.NewTaxonomy("NCBI Taxonomy", "taxon", "[[:digit:]]")

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

	root := taxonomy.Taxon("1")
	taxonomy.SetRoot(root)

	return taxonomy, nil
}
