package obitag2

import (
	"sort"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

type Obitag2Match struct {
	Taxon  *obitax.TaxNode
	Id     string
	Weight int
}
type Obitag2RefDB struct {
	Taxonomy *obitax.Taxonomy

	Full       *obiseq.BioSequenceSlice
	FullCounts *[]*obikmer.Table4mer
	FullTaxa   *obitax.TaxonSet

	Clusters      *obiseq.BioSequenceSlice
	ClusterCounts *[]*obikmer.Table4mer
	ClusterTaxa   *obitax.TaxonSet

	Families     *map[int]*obiseq.BioSequenceSlice
	FamilyCounts *map[int]*[]*obikmer.Table4mer
	FamilyTaxa   *map[int]obitax.TaxonSet

	ExactTaxid *map[string]*Obitag2Match
}

// MatchDistanceIndex returns the taxid, rank, and scientificName based on the given distance and distanceIdx.
//
// Parameters:
// - distance: The distance to match against the keys in distanceIdx.
// - distanceIdx: A map containing distances as keys and corresponding values in the format "taxid@rank@scientificName".
//
// Returns:
// - taxid: The taxid associated with the matched distance.
// - rank: The rank associated with the matched distance.
// - scientificName: The scientific name associated with the matched distance.
func MatchDistanceIndex(distance int, distanceIdx map[int]string) (int, string, string) {
	keys := maps.Keys(distanceIdx)
	slices.Sort(keys)

	i := sort.Search(len(keys), func(i int) bool {
		return distance <= keys[i]
	})

	var taxid int
	var rank string
	var scientificName string

	if i == len(keys) || distance > keys[len(keys)-1] {
		taxid = 1
		rank = "no rank"
		scientificName = "root"
	} else {
		parts := strings.Split(distanceIdx[keys[i]], "@")
		taxid, _ = strconv.Atoi(parts[0])
		rank = parts[1]
		scientificName = parts[2]
	}

	// log.Info("taxid:", taxid, " rank:", rank, " scientificName:", scientificName)

	return taxid, rank, scientificName
}

// FindClosests finds the closest bio sequence from a given sequence and a slice of reference sequences.
//
// Parameters:
// - sequence: the bio sequence to find the closest matches for.
// - references: a slice of reference sequences to compare against.
// - refcounts: a slice of reference sequence counts.
// - runExact: a boolean flag indicating whether to run an exact match.
//
// Returns:
// - bests: a slice of the closest bio sequences.
// - maxe: the maximum score.
// - bestId: the best ID.
// - bestmatch: the best match.
// - bestidxs: a slice of the best indexes.
func FindClosests(sequence *obiseq.BioSequence,
	references obiseq.BioSequenceSlice,
	refcounts []*obikmer.Table4mer,
	runExact bool) (obiseq.BioSequenceSlice, int, float64, string, []int) {

	var matrix []uint64

	seqwords := obikmer.Count4Mer(sequence, nil, nil)
	cw := make([]int, len(refcounts))

	for i, ref := range refcounts {
		cw[i] = obikmer.Common4Mer(seqwords, ref)
	}

	o := obiutils.Reverse(obiutils.IntOrder(cw), true)

	bests := obiseq.MakeBioSequenceSlice()
	//	bests = append(bests, references[o[0]])
	bestidxs := make([]int, 0)
	//	bestidxs = append(bestidxs, o[0])
	bestId := 0.0
	bestmatch := references[o[0]].Id()

	maxe := -1
	wordmin := 0

	// sss := make([]int, 0, 10000)
	// log.Warnf("%v\n%v\n%d", seqwords, refcounts[o[0]], obikmer.Common4Mer(seqwords, refcounts[o[0]]))

	for i, order := range o {
		ref := references[order]
		score := int(1e9)

		if cw[order] < wordmin || i > 1000 {
			break
		}
		lcs, alilength := -1, -1
		switch maxe {
		case 0:
			if obiutils.UnsafeStringFromBytes(sequence.Sequence()) == obiutils.UnsafeStringFromBytes(references[order].Sequence()) {
				score = 0
				alilength = sequence.Len()
				lcs = alilength
			}
		case 1:
			d, _, _, _ := obialign.D1Or0(sequence, references[order])
			if d >= 0 {
				score = d
				alilength = max(sequence.Len(), ref.Len())
				lcs = alilength - score
			}
		default:
			lcs, alilength = obialign.FastLCSScore(sequence, references[order], maxe, &matrix)
			if lcs >= 0 {
				score = alilength - lcs
			}
		}

		// log.Warnf("LCS : %d  ALILENGTH : %d score : %d cw : %d", lcs, alilength, score, cw[order])
		if lcs >= 0 {

			// sss = append(sss, cw[order], alilength-lcs)

			//
			// We have found a better candidate than never
			//
			if maxe == -1 || score < maxe {
				bests = bests[:0]       // Empty the best lists
				bestidxs = bestidxs[:0] //

				maxe = score // Memorize the best scores
				wordmin = max(0, max(sequence.Len(), ref.Len())-3-4*maxe)
				bestId = float64(lcs) / float64(alilength)
				bestmatch = ref.Id()
				// log.Warnln(sequence.Id(), ref.Id(), cw[order], maxe, bestId, wordmin)
			}

			if score == maxe {
				bests = append(bests, ref)
				bestidxs = append(bestidxs, order)
				id := float64(lcs) / float64(alilength)
				if id > bestId {
					bestId = id
					bestmatch = ref.Id()
				}
				// log.Println(ref.Id(), maxe, bestId,bestidxs)
			}

		}

	}
	// sequence.SetAttribute("maxe", sss)
	log.Debugln("Closest Match", sequence.Id(), maxe, bestId, references[bestidxs[0]].Id(), bestidxs, len(bests))
	return bests, maxe, bestId, bestmatch, bestidxs
}

func (db *Obitag2RefDB) BestConsensus(bests obiseq.BioSequenceSlice, differences int, slot string) *obitax.TaxNode {
	taxon := (*obitax.TaxNode)(nil)

	for _, best := range bests {
		idx := best.OBITagRefIndex(slot)
		if idx == nil {
			log.Fatalf("Reference database must be indexed first")
		}

		d := differences
		identification, ok := idx[d]
		found := false
		var parts []string

		/*
			Here is an horrible hack for xprize challence.
			With Euka01 the part[0] was equal to "" for at
			least a sequence consensus. Which is not normal.

			TO BE CHECKED AND CORRECTED

			The problem seems related to idx that doesn't have
			a 0 distance
		*/
		for !found && d >= 0 {
			for !ok && d >= 0 {
				identification, ok = idx[d]
				d--
			}

			parts = strings.Split(identification, "@")

			found = parts[0] != ""
			if !found {
				log.Debugln("Problem in identification line : ", best.Id(), "idx:", idx, "distance:", d)
				for !ok && d <= 1000 {
					identification, ok = idx[d]
					d++
				}

			}
		}

		match_taxid, err := strconv.Atoi(parts[0])

		if err != nil {
			log.Panicln("Cannot extract taxid from :", identification)
		}

		match_taxon, err := db.Taxonomy.Taxon(match_taxid)

		if err != nil {
			log.Panicln("Cannot find taxon corresponding to taxid :", match_taxid)
		}

		if taxon != nil {
			taxon, _ = taxon.LCA(match_taxon)
		} else {
			taxon = match_taxon
		}

	}

	return taxon
}

// Identify makes the taxonomic identification of a BioSequence.
//
// Parameters:
// - sequence: A pointer to a BioSequence to identify.
// - references: A BioSequenceSlice.
// - refcounts: A slice of pointers to Table4mer.
// - taxa: A TaxonSet.
// - taxo: A pointer to a Taxonomy.
// - runExact: A boolean value indicating whether to run exact matching.
//
// Returns:
// - A pointer to a BioSequence.
func Identify(sequence *obiseq.BioSequence,
	db *Obitag2RefDB,
	runExact bool) *obiseq.BioSequence {

	identity := 1.0
	differences := 0
	method := "exact match"
	weight := 0

	var bestmatch string
	var taxon *obitax.TaxNode

	exacttaxon, ok := (*db.ExactTaxid)[obiutils.UnsafeStringFromBytes(sequence.Sequence())]
	if ok {
		taxon = exacttaxon.Taxon
		bestmatch = exacttaxon.Id
		weight = exacttaxon.Weight
	} else {
		var bests obiseq.BioSequenceSlice
		bests, differences, identity, bestmatch, _ = FindClosests(sequence, *db.Clusters, *db.ClusterCounts, runExact)
		weight = bests.Len() /* Should be the some of the count */
		ftaxon := (*obitax.TaxNode)(nil)

		if identity >= 0.5 && differences >= 0 {
			ftaxon = db.BestConsensus(bests, differences, "obitag_ref_index")

			fam := ftaxon.TaxonAtRank("family")

			if fam != nil {
				ftaxid := fam.Taxid()
				bests, differences, identity, bestmatch, _ = FindClosests(
					sequence,
					*(*db.Families)[ftaxid],
					*(*db.FamilyCounts)[ftaxid],
					runExact,
				)
				sequence.SetAttribute("obitag_proposed_family", fam.ScientificName())
				ftaxon = db.BestConsensus(bests, differences, "reffamidx_in")

			}

		} else {
			// Cannot match any taxon
			ftaxon, _ = db.Taxonomy.Taxon(1)
			sequence.SetTaxid(1)
		}

		taxon = ftaxon
		method = "lcsfamlily"
	}

	sequence.SetTaxid(taxon.Taxid())

	sequence.SetAttribute("scientific_name", taxon.ScientificName())
	sequence.SetAttribute("obitag_rank", taxon.Rank())
	sequence.SetAttribute("obitag_bestid", identity)
	sequence.SetAttribute("obitag_bestmatch", bestmatch)
	sequence.SetAttribute("obitag_match_count", weight)
	sequence.SetAttribute("obitag_similarity_method", method)

	return sequence
}

func IdentifySeqWorker(db *Obitag2RefDB, runExact bool) obiseq.SeqWorker {
	return func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		return obiseq.BioSequenceSlice{Identify(sequence, db, runExact)}, nil
	}
}

func CLIAssignTaxonomy(iterator obiiter.IBioSequence,
	references obiseq.BioSequenceSlice,
	taxo *obitax.Taxonomy,
) obiiter.IBioSequence {

	refcounts := make(
		[]*obikmer.Table4mer,
		len(references))

	taxa := make(obitax.TaxonSet,
		len(references))

	familydb := make(map[int]*obiseq.BioSequenceSlice)
	familycounts := make(map[int]*[]*obikmer.Table4mer)
	familytaxa := make(map[int]obitax.TaxonSet)
	clusterdb := obiseq.MakeBioSequenceSlice()
	clustercounts := make([]*obikmer.Table4mer, 0)
	clustertaxa := make(obitax.TaxonSet, 0)
	exactmatch := make(map[string]*obiseq.BioSequenceSlice)

	buffer := make([]byte, 0, 1000)

	j := 0
	for i, seq := range references {
		refcounts[i] = obikmer.Count4Mer(seq, &buffer, nil)
		taxa[i], _ = taxo.Taxon(seq.Taxid())

		if is_centrer, _ := seq.GetBoolAttribute("reffamidx_clusterhead"); is_centrer {
			clusterdb = append(clusterdb, seq)
			clustercounts = append(clustercounts, refcounts[i])
			clustertaxa[j] = taxa[i]
			j++
		}

		family, _ := seq.GetIntAttribute("family_taxid")

		fs, ok := familydb[family]
		if !ok {
			fs = obiseq.NewBioSequenceSlice(0)
			familydb[family] = fs
		}

		*fs = append(*fs, seq)

		fc, ok := familycounts[family]

		if !ok {
			fci := make([]*obikmer.Table4mer, 0)
			fc = &fci
			familycounts[family] = fc
		}

		*fc = append(*fc, refcounts[i])

		ft, ok := familytaxa[family]

		if !ok {
			ft = make(obitax.TaxonSet, 0)
			familytaxa[family] = ft
		}

		ft[len(ft)] = taxa[i]

		seqstr := obiutils.UnsafeStringFromBytes(seq.Sequence())
		em, ok := exactmatch[seqstr]

		if !ok {
			em = obiseq.NewBioSequenceSlice()
			exactmatch[seqstr] = em
		}

		*em = append(*em, seq)

	}

	exacttaxid := make(map[string]*Obitag2Match, len(exactmatch))
	for seqstr, seqs := range exactmatch {
		var err error
		t, _ := taxo.Taxon((*seqs)[0].Taxid())
		w := (*seqs)[0].Count()
		lseqs := seqs.Len()

		for i := 1; i < lseqs; i++ {
			t2, _ := taxo.Taxon((*seqs)[i].Taxid())
			t, err = t.LCA(t2)

			if err != nil {
				log.Panic(err)
			}

			w += (*seqs)[i].Count()
		}

		exacttaxid[seqstr] = &Obitag2Match{
			Taxon:  t,
			Id:     (*seqs)[0].Id(),
			Weight: w,
		}
	}

	db := &Obitag2RefDB{
		Taxonomy: taxo,

		Full:       &references,
		FullCounts: &refcounts,
		FullTaxa:   &taxa,

		Clusters:      &clusterdb,
		ClusterCounts: &clustercounts,
		ClusterTaxa:   &clustertaxa,

		Families:     &familydb,
		FamilyCounts: &familycounts,
		FamilyTaxa:   &familytaxa,

		ExactTaxid: &exacttaxid,
	}

	worker := IdentifySeqWorker(db, CLIRunExact())

	return iterator.MakeIWorker(worker, false, obioptions.CLIParallelWorkers(), 0)
}
