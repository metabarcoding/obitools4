package main

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitaxformat"
)

func main() {

	obitaxformat.DetectTaxonomyFormat(os.Args[1])
}
