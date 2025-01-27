package obitax

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

// loadNodeTable reads a node table from the provided reader and populates the given taxonomy.
// It is an internal function and should not be called directly. It is part of the NCBI taxdump reader.
// The node table is expected to be in CSV format with a custom delimiter ('|') and comments
// starting with '#'. Each record in the table represents a taxon with its taxid, parent taxid,
// and rank.
//
// Parameters:
//   - reader: An io.Reader from which the node table is read.
//   - taxonomy: A pointer to an obitax.Taxonomy instance where the taxon data will be added.
//
// The function reads each record from the input, trims whitespace from the taxid, parent, and rank,
// and adds the taxon to the taxonomy. If an error occurs while adding a taxon, the function logs
// a fatal error and terminates the program.
func loadNodeTable(reader io.Reader, taxonomy *Taxonomy) {
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

// loadNameTable reads a name table from the provided reader and populates the given taxonomy.
// It is an internal function and should not be called directly. It is part of the NCBI taxdump reader.
// The name table is expected to be in a custom format with fields separated by the '|' character.
// Each record in the table represents a taxon with its taxid, name, and class name.
//
// Parameters:
//   - reader: An io.Reader from which the name table is read.
//   - taxonomy: A pointer to an obitax.Taxonomy instance where the taxon names will be set.
//   - onlysn: A boolean flag indicating whether to only process records with the class name "scientific name".
//
// Returns:
//
//	The number of taxon names successfully loaded into the taxonomy. If a line is too long, -1 is returned.
//	The function processes each line, trims whitespace from the taxid, name, and class name, and sets
//	the name in the taxonomy if the conditions are met.
func loadNameTable(reader io.Reader, taxonomy *Taxonomy, onlysn bool) int {
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

// loadMergedTable reads a merged table from the provided reader and populates the given taxonomy.
// It is an internal function and should not be called directly. It is part of the NCBI taxdump reader.
// The merged table is expected to be in CSV format with a custom delimiter ('|') and comments
// starting with '#'. Each record in the table represents a mapping between an old taxid and a new taxid.
//
// Parameters:
//   - reader: An io.Reader from which the merged table is read.
//   - taxonomy: A pointer to an obitax.Taxonomy instance where the alias mappings will be added.
//
// Returns:
//
//	The number of alias mappings successfully loaded into the taxonomy. The function processes
//	each record, trims whitespace from the old and new taxid, and adds the alias to the taxonomy.
func loadMergedTable(reader io.Reader, taxonomy *Taxonomy) int {
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
func LoadNCBITaxDump(directory string, onlysn bool) (*Taxonomy, error) {

	taxonomy := NewTaxonomy("NCBI Taxonomy", "taxon", obiutils.AsciiDigitSet)

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
