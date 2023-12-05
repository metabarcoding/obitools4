package obilandmark

import (
	"os"
	"sort"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obistats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obifind"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obirefidx"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

// MapOnLandmarkSequences performs sequence mapping on a given library of bio sequences.
//
// Computes for each sequence in the library a descriptor vector containing describing the sequence
// as the set of its distances to every landmark sequence.
//
// Parameters:
// - library: A slice of bio sequences to be mapped.
// - landmark_idx: A list of indices representing landmark sequences.
// - sizes: Optional argument specifying the number of workers to use.
//
// Returns:
// - seqworld: A matrix of float64 values representing the mapped coordinates.
func MapOnLandmarkSequences(library obiseq.BioSequenceSlice, landmark_idx []int, sizes ...int) obiutils.Matrix[float64] {
	nworkers := obioptions.CLIParallelWorkers()

	if len(sizes) > 0 {
		nworkers = sizes[0]
	}

	library_size := len(library)
	n_landmark := len(landmark_idx)
	todo := make(chan int, 0)

	seqworld := obiutils.Make2DArray[float64](library_size, n_landmark)

	pbopt := make([]progressbar.Option, 0, 5)
	pbopt = append(pbopt,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(15),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetDescription("[Sequence mapping]"),
	)

	bar := progressbar.NewOptions(library_size, pbopt...)

	waiting := sync.WaitGroup{}
	waiting.Add(nworkers)

	compute_coordinates := func() {
		buffer := make([]uint64, 1000)
		for i := range todo {
			seq := library[i]
			coord := seqworld[i]
			for j := 0; j < n_landmark; j++ {
				landmark := library[landmark_idx[j]]
				match, lalign := obialign.FastLCSScore(landmark, seq, -1, &buffer)
				coord[j] = float64(lalign - match)
			}
			bar.Add(1)
		}
		waiting.Done()
	}

	for i := 0; i < nworkers; i++ {
		go compute_coordinates()
	}

	for i := 0; i < library_size; i++ {
		todo <- i
	}

	close(todo)

	waiting.Wait()

	return seqworld
}

// CLISelectLandmarkSequences selects landmark sequences from the given iterator and assigns attributes to the sequences.
//
// The fonction annotate the input set of sequences with two or three attributes:
//   - 'landmark_id' indicating which sequence was selected and to which landmark it corresponds.
//   - 'landmark_coord' indicates the coordinates of the sequence.
//   - 'landmark_class' indicates to which landmark (landmark_id) the sequence is the closest.
//
// Parameters:
// - iterator: an object of type obiiter.IBioSequence representing the iterator to select landmark sequences from.
//
// Returns:
//   - an object of type obiiter.IBioSequence providing the input sequence annotated with their coordinates respectively to
//     each selected landmark sequences and with an attribute 'landmark_id' indicating which sequence was selected and to
//     which landmark it corresponds.
func CLISelectLandmarkSequences(iterator obiiter.IBioSequence) obiiter.IBioSequence {

	library := iterator.Load()

	library_size := len(library)
	n_landmark := CLINCenter()

	landmark_idx := obistats.SampleIntWithoutReplacement(n_landmark, library_size)
	log.Infof("Library contains %d sequence", len(library))

	var seqworld obiutils.Matrix[float64]

	for loop := 0; loop < 2; loop++ {
		sort.IntSlice(landmark_idx).Sort()
		log.Debugf("Selected indices : %v", landmark_idx)

		seqworld = MapOnLandmarkSequences(library, landmark_idx)
		initialCenters := obiutils.Make2DArray[float64](n_landmark, n_landmark)
		for i, seq_idx := range landmark_idx {
			initialCenters[i] = seqworld[seq_idx]
		}

		// classes, centers := obistats.Kmeans(&seqworld, n_landmark, &initialCenters)
		classifier := obistats.MakeKmeansClustering(&seqworld, n_landmark, obistats.DefaultRG())
		_, centers, inertia, converged := classifier.Run(1000, 0.001)
		intertia := classifier.Inertia()
		_, centers, inertia, converged := obistats.Kmeans(&seqworld, n_landmark, 0.001, &initialCenters)

		dist_centers := 0.0
		for i := 0; i < n_landmark; i++ {
			center := (*centers)[i]
			icenter := initialCenters[i]
			for j := 0; j < n_landmark; j++ {
				diff := center[j] - icenter[j]
				dist_centers += diff * diff
			}
		}

		landmark_idx = obistats.KmeansBestRepresentative(&seqworld, centers)
		log.Infof("Inertia: %f, Dist centers: %f, converged: %t", inertia, dist_centers, converged)

	}

	sort.IntSlice(landmark_idx).Sort()

	log.Debugf("Selected indices : %v", landmark_idx)
	seqworld = MapOnLandmarkSequences(library, landmark_idx)

	seq_landmark := make(map[int]int, n_landmark)
	for i, val := range landmark_idx {
		seq_landmark[val] = i
	}

	initialCenters := obiutils.Make2DArray[float64](n_landmark, n_landmark)
	for i, seq_idx := range landmark_idx {
		initialCenters[i] = seqworld[seq_idx]
	}

	classes := obistats.AssignToClass(&seqworld, &initialCenters)

	for i, seq := range library {
		ic, _ := obiutils.InterfaceToIntSlice(seqworld[i])
		seq.SetCoordinate(ic)
		seq.SetAttribute("landmark_class", classes[i])
		if i, ok := seq_landmark[i]; ok {
			seq.SetAttribute("landmark_id", i)
		}
	}

	if obifind.CLIHasSelectedTaxonomy() {
		taxo, err := obifind.CLILoadSelectedTaxonomy()
		if err != nil {
			log.Fatal(err)
		}

		taxa := make(obitax.TaxonSet, len(library))

		for i, seq := range library {
			taxa[i], err = taxo.Taxon(seq.Taxid())
			if err != nil {
				log.Fatal(err)
			}
		}

		pbopt := make([]progressbar.Option, 0, 5)
		pbopt = append(pbopt,
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionSetWidth(15),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetDescription("[Sequence Indexing]"),
		)

		bar := progressbar.NewOptions(len(library), pbopt...)

		for i, seq := range library {
			idx := obirefidx.GeomIndexSesquence(i, library, &taxa, taxo)
			seq.SetOBITagGeomRefIndex(idx)

			if i%10 == 0 {
				bar.Add(10)
			}
		}
	}

	return obiiter.IBatchOver(library, obioptions.CLIBatchSize())

}
