package obitag

import (
	"log"
	"math"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
)

// ExtractLandmarkSeqs extracts landmark sequences from the given BioSequenceSlice.
//
// The landmark sequences are extracted from the given BioSequenceSlice and put in a new BioSequenceSlice
// in the order corresponding to their landmark IDs.
//
// references: A pointer to a BioSequenceSlice containing the references.
// Returns: A pointer to a BioSequenceSlice containing the extracted landmark sequences.
func ExtractLandmarkSeqs(references *obiseq.BioSequenceSlice) *obiseq.BioSequenceSlice {
	landmarks := make(map[int]*obiseq.BioSequence, 100)

	for _, ref := range *references {
		if id := ref.GetLandmarkID(); id != -1 {
			landmarks[id] = ref
		}
	}

	ls := obiseq.NewBioSequenceSlice(len(landmarks))
	*ls = (*ls)[0:len(landmarks)]

	for k, l := range landmarks {
		(*ls)[k] = l
	}

	return ls
}

// ExtractTaxonSet extracts a set of taxa from the given references and taxonomy.
//
// If a reference sequence has a taxid absent from the taxonomy, the function will panic.
//
// The function takes two parameters:
// - references: a pointer to a BioSequenceSlice, which is a slice of BioSequence objects.
// - taxonomy: a pointer to a Taxonomy object.
//
// The function returns a pointer to a TaxonSet, which is a set of taxa.
func ExtractTaxonSet(references *obiseq.BioSequenceSlice, taxonomy *obitax.Taxonomy) *obitax.TaxonSet {
	var err error
	taxa := make(obitax.TaxonSet, len(*references))

	for i, ref := range *references {
		taxid := ref.Taxid()
		taxa[i], err = taxonomy.Taxon(taxid)
		if err != nil {
			log.Panicf("Taxid %d, for sequence %s not found in taxonomy", taxid, ref.Id())
		}
	}

	return &taxa
}

// MapOnLandmarkSequences calculates the coordinates of landmarks on a given sequence.
//
// It takes in three parameters:
//   - sequence: a pointer to a BioSequence object representing the sequence.
//   - landmarks: a pointer to a BioSequenceSlice object representing the landmarks.
//   - buffer: a pointer to a slice of uint64, used as a buffer for calculations.
//
// It returns a slice of integers representing the coordinates of the landmarks on the sequence.
func MapOnLandmarkSequences(sequence *obiseq.BioSequence, landmarks *obiseq.BioSequenceSlice, buffer *[]uint64) []int {

	coords := make([]int, len(*landmarks))

	for i, l := range *landmarks {
		lcs, length := obialign.FastLCSEGFScore(sequence, l, -1, buffer)
		coords[i] = length - lcs
	}

	return coords
}

// FindGeomClosest finds the closest geometric sequence in a given set of reference sequences to a query sequence.
//
// Parameters:
// - sequence: A pointer to a BioSequence object representing the query sequence.
// - landmarks: A pointer to a BioSequenceSlice object representing the landmarks.
// - references: A pointer to a BioSequenceSlice object representing the reference sequences.
// - buffer: A pointer to a slice of uint64 representing a buffer.
//
// Returns:
// - A pointer to a BioSequence object representing the closest sequence.
// - An int representing the minimum distance.
// - A float64 representing the best identity score.
// - An array of int representing the indices of the closest sequences.
// - A pointer to a BioSequenceSlice object representing the matched sequences.
func FindGeomClosest(sequence *obiseq.BioSequence,
	landmarks *obiseq.BioSequenceSlice,
	references *obiseq.BioSequenceSlice,
	buffer *[]uint64) (*obiseq.BioSequence, int, float64, []int, *obiseq.BioSequenceSlice) {

	min_dist := math.MaxInt64
	min_idx := make([]int, 0)

	query_location := MapOnLandmarkSequences(sequence, landmarks, buffer)

	for i, l := range *references {
		coord := l.GetCoordinate()
		if len(coord) == 0 {
			log.Panicf("Empty coordinate for reference sequence %s", l.Id())
		}
		dist := 0
		for j := 0; j < len(coord); j++ {
			diff := query_location[j] - coord[j]
			dist += diff * diff
		}

		if dist == min_dist {
			min_idx = append(min_idx, i)
		}
		if dist < min_dist {
			min_dist = dist
			min_idx = make([]int, 0)
			min_idx = append(min_idx, i)
		}
	}

	best_seq := (*references)[min_idx[0]]
	best_id := 0.0

	for _, i := range min_idx {
		seq := (*references)[i]
		lcs, length := obialign.FastLCSEGFScore(sequence, seq, -1, buffer)
		ident := float64(lcs) / float64(length)
		if ident > best_id {
			best_id = ident
			best_seq = seq
		}
	}

	matches := obiseq.MakeBioSequenceSlice(len(min_idx))
	matches = matches[0:len(min_idx)]
	for i, j := range min_idx {
		matches[i] = (*references)[j]
	}

	return best_seq, min_dist, best_id, query_location, &matches
}

func GeomIdentify(sequence *obiseq.BioSequence,
	landmarks *obiseq.BioSequenceSlice,
	references *obiseq.BioSequenceSlice,
	taxa *obitax.TaxonSet,
	taxo *obitax.Taxonomy,
	buffer *[]uint64) *obiseq.BioSequence {
	best_seq, min_dist, best_id, query_location, matches := FindGeomClosest(sequence, landmarks, references, buffer)

	taxon := (*obitax.TaxNode)(nil)
	var err error

	if best_id > 0.5 {
		taxid, _, _ := MatchDistanceIndex(min_dist, (*matches)[0].OBITagGeomRefIndex())
		taxon, _ = taxo.Taxon(taxid)
		for i := 1; i < len(*matches); i++ {
			taxid, _, _ := MatchDistanceIndex(min_dist, (*matches)[i].OBITagGeomRefIndex())
			newTaxon, _ := taxo.Taxon(taxid)
			taxon, err = newTaxon.LCA(taxon)
			if err != nil {
				log.Panicf("LCA error: %v", err)
			}
		}
		sequence.SetTaxid(taxon.Taxid())
	} else {
		taxon, _ = taxo.Taxon(1)
		sequence.SetTaxid(1)
	}

	sequence.SetAttribute("scientific_name", taxon.ScientificName())
	sequence.SetAttribute("obitag_rank", taxon.Rank())
	sequence.SetAttribute("obitag_bestid", best_id)
	sequence.SetAttribute("obitag_bestmatch", best_seq.Id())
	sequence.SetAttribute("obitag_min_dist", min_dist)
	sequence.SetAttribute("obitag_coord", query_location)
	sequence.SetAttribute("obitag_match_count", len(*matches))
	sequence.SetAttribute("obitag_similarity_method", "geometric")

	return sequence
}

func GeomIdentifySeqWorker(references *obiseq.BioSequenceSlice,
	taxo *obitax.Taxonomy) obiseq.SeqWorker {

	landmarks := ExtractLandmarkSeqs(references)
	taxa := ExtractTaxonSet(references, taxo)
	return func(sequence *obiseq.BioSequence) *obiseq.BioSequence {
		buffer := make([]uint64, 100)
		return GeomIdentify(sequence, landmarks, references, taxa, taxo, &buffer)
	}
}

func CLIGeomAssignTaxonomy(iterator obiiter.IBioSequence,
	references obiseq.BioSequenceSlice,
	taxo *obitax.Taxonomy,
) obiiter.IBioSequence {

	worker := GeomIdentifySeqWorker(&references, taxo)
	return iterator.MakeIWorker(worker, obioptions.CLIParallelWorkers(), 0)
}
