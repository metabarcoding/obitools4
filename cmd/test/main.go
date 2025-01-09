package main

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obifp"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"

	log "github.com/sirupsen/logrus"
)

func main() {

	one := obifp.OneUint[obifp.Uint128]()
	a, b := obifp.OneUint[obifp.Uint64]().LeftShift64(66, 0)
	log.Infof("one: %v, %v", a, b)
	shift := one.LeftShift(66)
	log.Infof("one: %v", shift)

	seq := obiseq.NewBioSequence("test", []byte("atcgggttccaacc"), "")

	kmermap := obikmer.NewKmerMap[obifp.Uint128](
		obiseq.BioSequenceSlice{
			seq,
		},
		7,
		true,
		10,
	)

	kmers := kmermap.NormalizedKmerSlice(seq, nil)

	for _, kmer := range kmers {
		println(kmermap.KmerAsString(kmer))
	}

}
