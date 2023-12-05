package obistats

import (
	"math"
	"sync"
	"time"

	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/sampleuv"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

func squareDist(a, b []float64) float64 {
	sum := 0.0
	for i := 0; i < len(a); i++ {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return sum
}

func DefaultRG() *rand.Rand {
	return rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
}

type KmeansClustering struct {
	data     *obiutils.Matrix[float64]
	rg       *rand.Rand
	centers  obiutils.Matrix[float64]
	icenters []int
	sizes    []int
	distmin  []float64
	classes  []int
}

func MakeKmeansClustering(data *obiutils.Matrix[float64], k int, rg *rand.Rand) *KmeansClustering {
	distmin := make([]float64, len(*data))
	for i := 0; i < len(distmin); i++ {
		distmin[i] = math.MaxFloat64
	}

	clustering := &KmeansClustering{
		data:     data,
		icenters: make([]int, 0, k),
		sizes:    make([]int, 0, k),
		centers:  make(obiutils.Matrix[float64], 0, k),
		distmin:  distmin,
		classes:  make([]int, len(*data)),
		rg:       rg,
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
func (clustering *KmeansClustering) AddACenter() {
	C := 0
	if clustering.K() == 0 {
		C = rand.Intn(clustering.N())
	} else {
		w := sampleuv.NewWeighted(clustering.distmin, clustering.rg)
		C, _ = w.Take()
	}
	clustering.icenters = append(clustering.icenters, C)
	clustering.sizes = append(clustering.sizes, 0)
	center := (*clustering.data)[C]
	clustering.centers = append(clustering.centers, center)

	n := clustering.N()

	for i := 0; i < n; i++ {
		d := squareDist((*clustering.data)[i], center)
		if d < clustering.distmin[i] {
			clustering.distmin[i] = d
		}
	}
}

// ResetEmptyCenters resets the empty centers in the KmeansClustering struct.
//
// It iterates over the centers and checks if their corresponding sizes are zero.
// If a center is empty, a new weighted sample is taken with the help of the distmin and rg variables.
// The new center is then assigned to the empty center index, and the sizes and centers arrays are updated accordingly.
// Finally, the function returns the number of empty centers that were reset.
func (clustering *KmeansClustering) ResetEmptyCenters() int {
	nreset := 0
	for i := 0; i < clustering.K(); i++ {
		if clustering.sizes[i] == 0 {
			w := sampleuv.NewWeighted(clustering.distmin, clustering.rg)
			C, _ := w.Take()
			clustering.icenters[i] = C
			clustering.centers[i] = (*clustering.data)[C]
			nreset++
		}
	}
	return nreset
}

// AssignToClass assigns each data point to a class based on the distance to the nearest center.
//
// This function does not take any parameters.
// It does not return anything.
func (clustering *KmeansClustering) AssignToClass() {
	var wg sync.WaitGroup
	var lock sync.Mutex

	for i := 0; i < clustering.K(); i++ {
		clustering.sizes[i] = 0
	}
	for i := 0; i < clustering.N(); i++ {
		clustering.distmin[i] = math.MaxFloat64
	}

	goroutine := func(i int) {
		defer wg.Done()
		dmin := math.MaxFloat64
		cmin := -1
		for j, center := range clustering.centers {
			dist := squareDist((*clustering.data)[i], center)
			if dist < dmin {
				dmin = dist
				cmin = j
			}
		}
		lock.Lock()
		clustering.classes[i] = cmin
		clustering.sizes[cmin]++
		clustering.distmin[i] = dmin
		lock.Unlock()
	}

	wg.Add(clustering.N())
	for i := 0; i < clustering.N(); i++ {
		go goroutine(i)
	}

	nreset := clustering.ResetEmptyCenters()

	if nreset > 0 {
		log.Warnf("Reset %d empty centers", nreset)
		clustering.AssignToClass()
	}
}

// ComputeCenters calculates the centers of the K-means clustering algorithm.
//
// It takes no parameters.
// It does not return any values.
func (clustering *KmeansClustering) ComputeCenters() {
	var wg sync.WaitGroup
	centers := clustering.centers
	data := clustering.data
	classes := clustering.classes
	k := clustering.K()

	// Goroutine code
	goroutine1 := func(centerIdx int) {
		defer wg.Done()
		for j, row := range *data {
			class := classes[j]
			if class == centerIdx {
				for l, val := range row {
					centers[centerIdx][l] += val
				}
			}
		}
	}

	for i := 0; i < k; i++ {
		wg.Add(1)
		go goroutine1(i)
	}

	wg.Wait()

	for i := range centers {
		for j := range centers[i] {
			centers[i][j] /= float64(clustering.sizes[i])
		}
	}

	goroutine2 := func(centerIdx int) {
		defer wg.Done()
		dkmin := math.MaxFloat64
		dki := -1
		center := centers[centerIdx]
		for j, row := range *data {
			if classes[j] == centerIdx {
				dist := squareDist(row, center)
				if dist < dkmin {
					dkmin = dist
					dki = j
				}
			}
		}
		clustering.icenters[centerIdx] = dki
		clustering.centers[centerIdx] = (*data)[dki]
	}

	for i := 0; i < k; i++ {
		wg.Add(1)
		go goroutine2(i)
	}

	wg.Wait()

}

func (clustering *KmeansClustering) Inertia() float64 {
	inertia := 0.0

	for i := 0; i < clustering.N(); i++ {
		inertia += clustering.distmin[i]
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
		clustering.AssignToClass()
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
