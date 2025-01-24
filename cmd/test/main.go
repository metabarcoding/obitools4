package main

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitaxformat"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

func main() {

	obitaxformat.DetectTaxonomyFormat(os.Args[1])
	println(obiutils.RemoveAllExt("toto/tutu/test.txt"))
	println(obiutils.Basename("toto/tutu/test.txt"))

}
