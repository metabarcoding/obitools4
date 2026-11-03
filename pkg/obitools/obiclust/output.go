package obiclust

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiitercsv"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// Annotate writes the result of the clustering on the sequences and returns
// them in the order sumaclust writes them: the groups of identical sequences
// in their sorted order, each group beginning with its head.
//
// Every sequence receives
//   - cluster: the identifier of the center of its cluster;
//   - cluster_score: the score with this center;
//   - cluster_center: true for the center only;
//   - cluster_weight: the cumulated count of the sequences of the cluster.
//
// A sequence with IUPAC letters that matches no center belongs to no cluster:
// it only gets cluster_center=false.
func (res *Result) Annotate() obiseq.BioSequenceSlice {
	recs, groups, cl := res.recs, res.groups, res.clustered
	weight := make([]int, len(groups))
	for g := range groups {
		if cl.center[g] >= 0 {
			weight[cl.center[g]] += groups[g].count
		}
	}

	out := make(obiseq.BioSequenceSlice, 0, len(recs))
	for g, grp := range groups {
		ctr := cl.center[g]

		for r := grp.start; r < grp.end; r++ {
			s := recs[r].seq
			if ctr < 0 {
				// a sequence with IUPAC letters that matches no center
				s.RemoveAttribute("cluster")
				s.RemoveAttribute("cluster_score")
				s.RemoveAttribute("cluster_weight")
				s.SetAttribute("cluster_center", false)
				out = append(out, s)
				continue
			}
			centerID := recs[groups[ctr].start].seq.Id()
			s.SetAttribute("cluster", centerID)
			s.SetAttribute("cluster_score", cl.score[g])
			s.SetAttribute("cluster_center", ctr == g && r == grp.start)
			s.SetAttribute("cluster_weight", weight[ctr])
			out = append(out, s)
		}
	}
	return out
}

// OTUMapIterator describes the clusters as a CSV table, one row per sequence:
//   - id: the identifier of the sequence;
//   - cluster: the identifier of the center of its cluster;
//   - cluster_center: true for the center only;
//   - cluster_score: the score with this center;
//   - count: the count of the sequence.
//
// The rows are grouped by cluster, in the order the clusters were created. In
// a cluster, the center comes first, followed by the sequences identical to
// it, then by the other members, each followed by its identical sequences.
// The sequences without cluster (see Annotate) come last, with no cluster nor
// score.
func (res *Result) OTUMapIterator(source string, batchSize int) *obiitercsv.ICSVRecord {
	recs, groups, cl := res.recs, res.groups, res.clustered

	iterator := obiitercsv.NewICSVRecord()
	iterator.Add(1)
	for _, field := range []string{"id", "cluster", "cluster_center", "cluster_score", "count"} {
		iterator.AppendField(field)
	}

	members := make(map[int][]int, len(cl.centers))
	var orphans []int
	for g := range groups {
		if ctr := cl.center[g]; ctr >= 0 {
			members[ctr] = append(members[ctr], g)
		} else {
			orphans = append(orphans, g)
		}
	}

	go func() {
		iterator.WaitAndClose()
	}()

	go func() {
		order := 0
		batch := make([]obiitercsv.CSVRecord, 0, batchSize)
		add := func(g int) {
			ctr := cl.center[g]
			for r := groups[g].start; r < groups[g].end; r++ {
				row := obiitercsv.CSVRecord{
					"id":             recs[r].seq.Id(),
					"cluster_center": ctr == g && r == groups[g].start,
					"count":          recs[r].count,
				}
				if ctr >= 0 {
					row["cluster"] = recs[groups[ctr].start].seq.Id()
					row["cluster_score"] = cl.score[g]
				}
				batch = append(batch, row)
				if len(batch) == batchSize {
					iterator.Push(obiitercsv.MakeCSVRecordBatch(source, order, batch))
					order++
					batch = make([]obiitercsv.CSVRecord, 0, batchSize)
				}
			}
		}
		for _, ctr := range cl.centers {
			for _, g := range members[ctr] {
				add(g)
			}
		}
		for _, g := range orphans {
			add(g)
		}
		if len(batch) > 0 || order == 0 {
			iterator.Push(obiitercsv.MakeCSVRecordBatch(source, order, batch))
		}
		iterator.Done()
	}()

	return iterator
}
