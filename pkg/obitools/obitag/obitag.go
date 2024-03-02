package obitag

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
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obirefidx"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

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

	for _, order := range o {
		ref := references[order]
		score := int(1e9)

		if maxe != -1 {
			wordmin = obiutils.MaxInt(sequence.Len(), ref.Len()) - 3 - 4*maxe
		}

		if cw[order] < wordmin {
			break
		}

		lcs, alilength := -1, -1
		if maxe == 0 || maxe == 1 {
			d, _, _, _ := obialign.D1Or0(sequence, references[order])
			if d >= 0 {
				score = d
				alilength = obiutils.MaxInt(sequence.Len(), ref.Len())
				lcs = alilength - score
			}
		} else {
			lcs, alilength = obialign.FastLCSScore(sequence, references[order], maxe, &matrix)
			if lcs >= 0 {
				score = alilength - lcs
			}
		}

		if maxe == -1 || score < maxe {
			bests = bests[:0]
			bestidxs = bestidxs[:0]
			maxe = score
			bestId = float64(lcs) / float64(alilength)
			// log.Println(ref.Id(), maxe, bestId,bestidxs)
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

	// log.Debugln("Closest Match", sequence.Id(), maxe, bestId, bestidxs, len(bests))
	return bests, maxe, bestId, bestmatch, bestidxs
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
	references obiseq.BioSequenceSlice,
	refcounts []*obikmer.Table4mer,
	taxa obitax.TaxonSet,
	taxo *obitax.Taxonomy,
	runExact bool) *obiseq.BioSequence {
	bests, differences, identity, bestmatch, seqidxs := FindClosests(sequence, references, refcounts, runExact)
	taxon := (*obitax.TaxNode)(nil)

	if identity >= 0.5 && differences >= 0 {
		newidx := 0
		for i, best := range bests {
			idx := best.OBITagRefIndex()
			if idx == nil {
				// log.Debugln("Need of indexing")
				newidx++
				idx = obirefidx.IndexSequence(seqidxs[i], references, &refcounts, &taxa, taxo)
				references[seqidxs[i]].SetOBITagRefIndex(idx)
				log.Debugln(references[seqidxs[i]].Id(), idx)
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

			match_taxon, err := taxo.Taxon(match_taxid)

			if err != nil {
				log.Panicln("Cannot find taxon corresponding to taxid :", match_taxid)
			}

			if taxon != nil {
				taxon, _ = taxon.LCA(match_taxon)
			} else {
				taxon = match_taxon
			}

		}
		log.Debugln(sequence.Id(), "Best matches:", len(bests), "New index:", newidx)

		sequence.SetTaxid(taxon.Taxid())

	} else {
		taxon, _ = taxo.Taxon(1)
		sequence.SetTaxid(1)
	}

	sequence.SetAttribute("scientific_name", taxon.ScientificName())
	sequence.SetAttribute("obitag_rank", taxon.Rank())
	sequence.SetAttribute("obitag_bestid", identity)
	sequence.SetAttribute("obitag_bestmatch", bestmatch)
	sequence.SetAttribute("obitag_match_count", len(bests))
	sequence.SetAttribute("obitag_similarity_method", "lcs")

	return sequence
}

func IdentifySeqWorker(references obiseq.BioSequenceSlice,
	refcounts []*obikmer.Table4mer,
	taxa obitax.TaxonSet,
	taxo *obitax.Taxonomy,
	runExact bool) obiseq.SeqWorker {
	return func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		return obiseq.BioSequenceSlice{Identify(sequence, references, refcounts, taxa, taxo, runExact)}, nil
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

	buffer := make([]byte, 0, 1000)

	for i, seq := range references {
		refcounts[i] = obikmer.Count4Mer(seq, &buffer, nil)
		taxa[i], _ = taxo.Taxon(seq.Taxid())
	}

	worker := IdentifySeqWorker(references, refcounts, taxa, taxo, CLIRunExact())

	return iterator.MakeIWorker(worker, false, obioptions.CLIParallelWorkers(), 0)
}
