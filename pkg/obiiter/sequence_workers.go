package obiiter

// func MakeSetAttributeWorker(rank string) obiiter.SeqWorker {

// 	if !goutils.Contains(taxonomy.RankList(), rank) {
// 		log.Fatalf("%s is not a valid rank (allowed ranks are %v)",
// 			rank,
// 			taxonomy.RankList())
// 	}

// 	w := func(sequence *obiseq.BioSequence) *obiseq.BioSequence {
// 		taxonomy.SetTaxonAtRank(sequence, rank)
// 		return sequence
// 	}

// 	return w
// }