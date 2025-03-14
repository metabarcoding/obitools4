package main

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiformats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

func main() {

	obiformats.DetectTaxonomyFormat(os.Args[1])
	println(obiutils.RemoveAllExt("toto/tutu/test.txt"))
	println(obiutils.Basename("toto/tutu/test.txt"))

}
