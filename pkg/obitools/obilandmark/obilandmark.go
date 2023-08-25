package obilandmark

import (
	"math"
	"os"
	"sort"
	"sync"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obialign"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obistats"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiutils"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

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

func CLISelectLandmarkSequences(iterator obiiter.IBioSequence) obiiter.IBioSequence {

	library := iterator.Load()

	library_size := len(library)
	n_landmark := NCenter()

	landmark_idx := obistats.SampleIntWithoutReplacemant(n_landmark, library_size)
	log.Infof("Library contains %d sequence", len(library))

	var seqworld obiutils.Matrix[float64]

	for loop := 0; loop < 5; loop++ {
		sort.IntSlice(landmark_idx).Sort()
		log.Infof("Selected indices : %v", landmark_idx)

		seqworld = MapOnLandmarkSequences(library, landmark_idx)
		initialCenters := obiutils.Make2DArray[float64](n_landmark, n_landmark)
		for i, seq_idx := range landmark_idx {
			initialCenters[i] = seqworld[seq_idx]
		}

		// classes, centers := obistats.Kmeans(&seqworld, n_landmark, &initialCenters)
		_, centers, inertia, converged := obistats.Kmeans(&seqworld, n_landmark, 0.001, &initialCenters)

		dist_centers := 0.0
		for i := 0; i < n_landmark; i++ {
			for j := 0; j < n_landmark; j++ {
				dist_centers += math.Pow((*centers)[i][j]-initialCenters[i][j], 2)
			}
		}

		landmark_idx = obistats.KmeansBestRepresentative(&seqworld, centers)
		log.Infof("Inertia: %f, Dist centers: %f, converged: %t", inertia, dist_centers, converged)

	}

	sort.IntSlice(landmark_idx).Sort()

	log.Infof("Selected indices : %v", landmark_idx)
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
		seq.SetAttribute("landmark_coord", seqworld[i])
		seq.SetAttribute("landmark_class", classes[i])
		if i, ok := seq_landmark[i]; ok {
			seq.SetAttribute("landmark_id", i)
		}
	}

	return obiiter.IBatchOver(library, obioptions.CLIBatchSize())

}
