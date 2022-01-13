package obitax

func (taxonomy *Taxonomy) RankList() []string {
	ranks := make([]string, 0, 30)
	mranks := make(map[string]bool)

	for _, t := range *taxonomy.nodes {
		mranks[t.rank] = true
	}

	for r := range mranks {
		ranks = append(ranks, r)
	}

	return ranks
}
