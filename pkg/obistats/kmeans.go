package obistats

import (
	"math"
	"sync"
	"time"

	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/sampleuv"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obilog"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

// SquareDist calculates the squared Euclidean distance between
// two vectors 'a' and 'b'.
//
// 'a' and 'b' are slices of float64 or int values representing
// coordinate points in space. It is assumed that both slices
// have the same length.
// Returns the calculated squared distance as a float64.
func SquareDist[T float64 | int](a, b []T) T {
	sum := T(0)
	for i, v := range a {
		diff := v - b[i]
		sum += diff * diff
	}
	return sum
}

// EuclideanDist calculates the Euclidean distance between
// two vectors represented as slices of float64.
//
// `a` and `b` are slices of float64 where each element of `a`
// is paired with the corresponding element of `b`.
// Returns the squared sum of the differences.
func EuclideanDist[T float64 | int](a, b []T) float64 {
	return math.Sqrt(float64(SquareDist(a, b)))
}

// DefaultRG creates and returns a new instance of *rand.Rand.
//
// No parameters.
// Returns *rand.Rand which is a pointer to a new random number
// generator, seeded with the current time in nanoseconds.
func DefaultRG() *rand.Rand {
	return rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
}

type KmeansClustering struct {
	data     *obiutils.Matrix[float64] // data matrix    						dimensions: n x p
	distmin  []float64                 // distance to closest center			dimension: n
	classes  []int                     // class of each data point				dimension: n
	rg       *rand.Rand                // random number generator
	centers  obiutils.Matrix[float64]  // centers coordinates 					dimensions: k x p
	icenters []int                     // indices of centers					dimension: k
	sizes    []int                     // number of elements in each cluster	dimension: k
}

// MakeKmeansClustering initializes a KmeansClustering with the
// provided matrix data, number of clusters k, and random number
// generator rg.
//
// data is a pointer to a Matrix of float64 representing the dataset,
// k is the number of desired clusters, and rg is a pointer to a
// random number generator used in the clustering process.
// Returns a pointer to the initialized KmeansClustering structure.
func MakeKmeansClustering(data *obiutils.Matrix[float64], k int, rg *rand.Rand) *KmeansClustering {
	distmin := make([]float64, len(*data))
	classes := make([]int, len(*data))
	for i := 0; i < len(distmin); i++ {
		distmin[i] = math.MaxFloat64
		classes[i] = -1
	}

	clustering := &KmeansClustering{
		data:     data,
		distmin:  distmin,
		classes:  classes,
		rg:       rg,
		centers:  make(obiutils.Matrix[float64], 0, k),
		icenters: make([]int, 0, k),
		sizes:    make([]int, 0, k),
	}

	for i := 0; i < k; i++ {
		clustering.AddACenter()
	}

	return clustering
}

// K returns the number of clusters in the K-means clustering algorithm.
//
// No parameters.
// Returns an integer.
func (clustering *KmeansClustering) K() int {
	return len(clustering.icenters)
}

// N returns the size of the dataset in the KmeansClustering instance.
//
// It does not take any parameters.
// The return type is an integer.
func (clustering *KmeansClustering) N() int {
	return len(*clustering.data)
}

// Dimension returns the dimension of the KmeansClustering data.
//
// No parameters.
// Returns an integer representing the dimension of the data.
func (clustering *KmeansClustering) Dimension() int {
	return len((*clustering.data)[0])
}

// SetCenterTo sets the center of a specific cluster to a given data point index.
//
// Parameters:
// - k: the index of the cluster, if k=-1, a new center is added
// - i: the index of the data point
// - reset: a boolean indicating whether to reset the distances to the nearest center
//          for points previously assigned to this center
//
// No return value.

func (clustering *KmeansClustering) SetCenterTo(k, i int, reset bool) {
	N := clustering.N()
	K := clustering.K()
	center := (*clustering.data)[i]

	if k >= 0 {
		clustering.icenters[k] = i
		clustering.sizes[k] = 0
		clustering.centers[k] = center

		if reset {
			// Recompute distances to the nearest center for points
			// previously assigned to this center
			K := clustering.K()
			for j := 0; j < N; j++ {
				if clustering.classes[j] == k {
					clustering.distmin[j] = math.MaxFloat64
					for l := 1; l < K; l++ {
						dist := EuclideanDist((*clustering.data)[j], clustering.centers[l])
						if dist < clustering.distmin[j] {
							clustering.distmin[j] = dist
							clustering.classes[j] = l
						}
					}
					clustering.sizes[clustering.classes[j]]++
				}
			}
		}

	} else {
		clustering.icenters = append(clustering.icenters, i)
		clustering.sizes = append(clustering.sizes, 0)
		clustering.centers = append(clustering.centers, center)
		k = K
		K++
	}

	for j := 0; j < clustering.N(); j++ {
		dist := EuclideanDist((*clustering.data)[j], center)
		if dist < clustering.distmin[j] {
			if C := clustering.classes[j]; C >= 0 {
				clustering.sizes[C]--
			}
			clustering.distmin[j] = dist
			clustering.classes[j] = k
			clustering.sizes[k]++
		}
	}

}

// AddACenter adds a new center to the KmeansClustering.
//
// If there are no centers, it randomly selects a new center.
// If there are existing centers, it selects a new center with
// probability proportional to its distance from the nearest
// center. The center is then added to the clustering.
func (clustering *KmeansClustering) AddACenter() {
	k := clustering.K()
	C := 0

	if k == 0 {
		// if there are no centers yet, draw a sample as the first center
		C = rand.Intn(clustering.N())
	} else {
		// otherwise, draw a sample with a probability proportional
		// to its closest distance to a center
		w := sampleuv.NewWeighted(clustering.distmin, clustering.rg)
		C, _ = w.Take()
	}

	clustering.SetCenterTo(-1, C, false)
}

// ResetEmptyCenters reinitializes any centers in a KmeansClustering
// that have no assigned points.
//
// This method iterates over the centers and uses a weighted sampling
// to reset centers with a size of zero.
// Returns the number of centers that were reset.
func (clustering *KmeansClustering) ResetEmptyCenters() int {
	nreset := 0
	for i := 0; i < clustering.K(); i++ {
		if clustering.sizes[i] == 0 {
			w := sampleuv.NewWeighted(clustering.distmin, clustering.rg)
			C, _ := w.Take()
			clustering.SetCenterTo(i, C, false)
			nreset++
		}
	}
	return nreset
}

// ClosestPoint finds the index of the closest point in the
// clustering to the given coordinates.
//
// coordinates is a slice of float64 representing the point.
// Returns the index of the closest point as an int.
func (clustering *KmeansClustering) ClosestPoint(coordinates []float64) int {
	N := clustering.N()
	distmin := math.MaxFloat64
	C := -1
	for i := 0; i < N; i++ {
		dist := EuclideanDist((*clustering.data)[i], coordinates)
		if dist < distmin {
			distmin = dist
			C = i
		}
	}
	return C
}

// AssignToClass assigns each data point in the dataset to the nearest
// center (class) in a K-means clustering algorithm.
//
// Handles the reinitialization of empty centers after the assignment.
// No return values.
func (clustering *KmeansClustering) AssignToClass() {
	var wg sync.WaitGroup
	var lock sync.Mutex

	// initialize the number of points in each class
	for i := 0; i < clustering.K(); i++ {
		clustering.sizes[i] = 0
	}

	goroutine := func(i int) {
		defer wg.Done()
		dmin := math.MaxFloat64
		cmin := -1
		for j, center := range clustering.centers {
			dist := EuclideanDist((*clustering.data)[i], center)
			if dist < dmin {
				dmin = dist
				cmin = j
			}
		}

		clustering.classes[i] = cmin
		clustering.distmin[i] = dmin

		lock.Lock()
		clustering.sizes[cmin]++
		lock.Unlock()
	}

	wg.Add(clustering.N())
	for i := 0; i < clustering.N(); i++ {
		go goroutine(i)
	}

	wg.Wait()

	nreset := clustering.ResetEmptyCenters()

	if nreset > 0 {
		obilog.Warnf("Reseted %d empty centers", nreset)
	}
}

// SetCentersTo assigns new centers in the KmeansClustering
// structure given a slice of indices.
//
// The indices parameter is a slice of integers that
// corresponds to the new indices of the cluster centers in
// the dataset. It panics if any index is out of bounds.
// This method does not return any value.
func (clustering *KmeansClustering) SetCentersTo(indices []int) {
	for _, v := range indices {
		if v < 0 || v >= clustering.N() {
			log.Fatalf("Invalid center index: %d", v)
		}
	}

	clustering.icenters = indices
	K := len(indices)

	for i := 0; i < K; i++ {
		clustering.centers[i] = (*clustering.data)[indices[i]]
	}

	clustering.AssignToClass()

}

// ComputeCenters calculates the centers of the K-means clustering algorithm.
//
// This method call AssignToClass() after computing the centers to ensure coherence
// of the clustering data structure.
//
// It takes no parameters.
// It does not return any values.
func (clustering *KmeansClustering) ComputeCenters() {
	var wg sync.WaitGroup
	centers := clustering.centers
	data := clustering.data
	classes := clustering.classes
	K := clustering.K()

	// compute the location of center of class centerIdx
	// as the point in the data the closest to the
	// center of class centerIdx
	newCenter := func(centerIdx int) {
		defer wg.Done()

		center := make([]float64, clustering.Dimension())

		for j := range center {
			center[j] = 0
		}

		for j, row := range *data {
			if classes[j] == centerIdx {
				for l, val := range row {
					center[l] += val
				}
			}
		}

		for j := range centers[centerIdx] {
			center[j] /= float64(clustering.sizes[centerIdx])
		}

		C := clustering.ClosestPoint(center)

		centers[centerIdx] = (*data)[C]
		clustering.icenters[centerIdx] = C
	}

	for i := 0; i < K; i++ {
		wg.Add(1)
		go newCenter(i)
	}

	wg.Wait()

	clustering.AssignToClass()

}

func (clustering *KmeansClustering) Inertia() float64 {
	inertia := 0.0

	for i := 0; i < clustering.N(); i++ {
		inertia += clustering.distmin[i] * clustering.distmin[i]
	}
	return inertia
}

func (clustering *KmeansClustering) Centers() obiutils.Matrix[float64] {
	return clustering.centers
}

func (clustering *KmeansClustering) CentersIndices() []int {
	return clustering.icenters
}

func (clustering *KmeansClustering) Sizes() []int {
	return clustering.sizes
}

func (clustering *KmeansClustering) Classes() []int {
	return clustering.classes
}

func (clustering *KmeansClustering) Run(max_cycle int, threshold float64) bool {
	prev := math.MaxFloat64
	newI := clustering.Inertia()
	for i := 0; i < max_cycle && (prev-newI) > threshold; i++ {
		prev = newI
		clustering.ComputeCenters()
		newI = clustering.Inertia()
	}

	return (prev - newI) <= threshold
}

// // Kmeans performs the K-means clustering algorithm on the given data.

// // if centers and *center is not nil, centers is considered as initialized
// // and the number of classes (k) is set to the number of rows in centers.
// // overwise, the number of classes is defined by the value of k.

// // Parameters:
// // - data: A pointer to a Matrix[float64] that represents the input data.
// // - k: An integer that specifies the number of clusters to create.
// // - threshold: A float64 value that determines the convergence threshold.
// // - centers: A pointer to a Matrix[float64] that represents the initial cluster centers.

// // Returns:
// // - classes: A slice of integers that assigns each data point to a cluster.
// // - centers: A pointer to a Matrix[float64] that contains the final cluster centers.
// // - inertia: A float64 value that represents the overall inertia of the clustering.
// // - converged: A boolean value indicating whether the algorithm converged.
// func Kmeans(data *obiutils.Matrix[float64],
// 	k int,
// 	threshold float64,
// 	centers *obiutils.Matrix[float64]) ([]int, *obiutils.Matrix[float64], float64, bool) {
// 	if centers == nil || *centers == nil {
// 		*centers = obiutils.Make2DArray[float64](k, len((*data)[0]))
// 		center_ids := SampleIntWithoutReplacement(k, len(*data))
// 		for i, id := range center_ids {
// 			(*centers)[i] = (*data)[id]
// 		}
// 	} else {
// 		k = len(*centers)
// 	}

// 	classes := AssignToClass(data, centers)
// 	centers = ComputeCenters(data, k, classes)
// 	inertia := ComputeInertia(data, classes, centers)
// 	delta := threshold * 100.0
// 	for i := 0; i < 100 && delta > threshold; i++ {
// 		classes = AssignToClass(data, centers)
// 		centers = ComputeCenters(data, k, classes)
// 		newi := ComputeInertia(data, classes, centers)
// 		delta = inertia - newi
// 		inertia = newi
// 		log.Debugf("Inertia: %f, delta: %f", inertia, delta)
// 	}

// 	return classes, centers, inertia, delta < threshold
// }

// // KmeansBestRepresentative finds the best representative among the data point of each cluster in parallel.
// //
// // It takes a matrix of data points and a matrix of centers as input.
// // The best representative is the data point that is closest to the center of the cluster.
// // Returns an array of integers containing the index of the best representative for each cluster.
// func KmeansBestRepresentative(data *obiutils.Matrix[float64], centers *obiutils.Matrix[float64]) []int {
// 	bestRepresentative := make([]int, len(*centers))

// 	var wg sync.WaitGroup
// 	wg.Add(len(*centers))

// 	for j, center := range *centers {
// 		go func(j int, center []float64) {
// 			defer wg.Done()

// 			bestDistToCenter := math.MaxFloat64
// 			best := -1

// 			for i, row := range *data {
// 				dist := 0.0
// 				for d, val := range row {
// 					diff := val - center[d]
// 					dist += diff * diff
// 				}
// 				if dist < bestDistToCenter {
// 					bestDistToCenter = dist
// 					best = i
// 				}
// 			}

// 			if best == -1 {
// 				log.Fatalf("No representative found for cluster %d", j)
// 			}

// 			bestRepresentative[j] = best
// 		}(j, center)
// 	}

// 	wg.Wait()

// 	return bestRepresentative
// }
