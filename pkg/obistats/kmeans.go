package obistats

import (
	"math"
	"sync"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

// AssignToClass applies the nearest neighbor algorithm to assign data points to classes.
//
// Parameters:
// - data: a 2D slice of float64 representing the data points to be assigned.
// - centers: a 2D slice of float64 representing the center points for each class.
//
// Return:
// - classes: a slice of int representing the assigned class for each data point.
func AssignToClass(data, centers *obiutils.Matrix[float64]) []int {
	classes := make([]int, len(*data))
	numData := len(*data)
	numCenters := len(*centers)

	var wg sync.WaitGroup
	wg.Add(numData)

	for i := 0; i < numData; i++ {
		go func(i int) {
			defer wg.Done()
			minDist := math.MaxFloat64
			minDistIndex := -1
			rowData := (*data)[i]

			for j := 0; j < numCenters; j++ {
				centerData := (*centers)[j]
				dist := 0.0

				for d, val := range rowData {
					diff := val - centerData[d]
					dist += diff * diff
				}

				if dist < minDist {
					minDist = dist
					minDistIndex = j
				}
			}

			classes[i] = minDistIndex
		}(i)
	}

	wg.Wait()

	return classes
}

// ComputeCenters calculates the centers of clusters for a given data set.
//
// Parameters:
// - data: a pointer to a matrix of float64 values representing the data set.
// - k: an integer representing the number of clusters.
// - classes: a slice of integers representing the assigned cluster for each data point.
//
// Returns:
// - centers: a pointer to a matrix of float64 values representing the centers of the clusters.
// ComputeCenters calculates the centers of clusters for a given data set.
//
// Parameters:
// - data: a pointer to a matrix of float64 values representing the data set.
// - k: an integer representing the number of clusters.
// - classes: a slice of integers representing the assigned cluster for each data point.
//
// Returns:
// - centers: a pointer to a matrix of float64 values representing the centers of the clusters.
func ComputeCenters(data *obiutils.Matrix[float64], k int, classes []int) *obiutils.Matrix[float64] {
	centers := obiutils.Make2DArray[float64](k, len((*data)[0]))
	centers.Init(0.0)
	ns := make([]int, k)

	var wg sync.WaitGroup

	for i := range ns {
		ns[i] = 0
	}

	// Goroutine code
	goroutine := func(centerIdx int) {
		defer wg.Done()
		for j, row := range *data {
			class := classes[j]
			if class == centerIdx {
				ns[centerIdx]++
				for l, val := range row {
					centers[centerIdx][l] += val
				}
			}
		}
	}

	for i := 0; i < k; i++ {
		wg.Add(1)
		go goroutine(i)
	}

	wg.Wait()

	for i := range centers {
		for j := range centers[i] {
			centers[i][j] /= float64(ns[i])
		}
	}

	return &centers
}

// ComputeInertia computes the inertia of the given data and centers in parallel.
//
// Parameters:
// - data: A pointer to a Matrix of float64 representing the data.
// - classes: A slice of int representing the class labels for each data point.
// - centers: A pointer to a Matrix of float64 representing the centers.
//
// Return type:
// - float64: The computed inertia.
func ComputeInertia(data *obiutils.Matrix[float64], classes []int, centers *obiutils.Matrix[float64]) float64 {
	inertia := make(chan float64)
	numRows := len(*data)
	wg := sync.WaitGroup{}
	wg.Add(numRows)

	for i := 0; i < numRows; i++ {
		go func(i int) {
			defer wg.Done()
			row := (*data)[i]
			class := classes[i]
			center := (*centers)[class]
			inertiaLocal := 0.0
			for j, val := range row {
				diff := val - center[j]
				inertiaLocal += diff * diff
			}
			inertia <- inertiaLocal
		}(i)
	}

	go func() {
		wg.Wait()
		close(inertia)
	}()

	totalInertia := 0.0
	for localInertia := range inertia {
		totalInertia += localInertia
	}

	return totalInertia
}

// Kmeans performs the K-means clustering algorithm on the given data.
//
// if centers and *center is not nil, centers is considered as initialized
// and the number of classes (k) is set to the number of rows in centers.
// overwise, the number of classes is defined by the value of k.
//
// Parameters:
// - data: A pointer to a Matrix[float64] that represents the input data.
// - k: An integer that specifies the number of clusters to create.
// - threshold: A float64 value that determines the convergence threshold.
// - centers: A pointer to a Matrix[float64] that represents the initial cluster centers.
//
// Returns:
// - classes: A slice of integers that assigns each data point to a cluster.
// - centers: A pointer to a Matrix[float64] that contains the final cluster centers.
// - inertia: A float64 value that represents the overall inertia of the clustering.
// - converged: A boolean value indicating whether the algorithm converged.
func Kmeans(data *obiutils.Matrix[float64],
	k int,
	threshold float64,
	centers *obiutils.Matrix[float64]) ([]int, *obiutils.Matrix[float64], float64, bool) {
	if centers == nil || *centers == nil {
		*centers = obiutils.Make2DArray[float64](k, len((*data)[0]))
		center_ids := SampleIntWithoutReplacement(k, len(*data))
		for i, id := range center_ids {
			(*centers)[i] = (*data)[id]
		}
	} else {
		k = len(*centers)
	}

	classes := AssignToClass(data, centers)
	centers = ComputeCenters(data, k, classes)
	inertia := ComputeInertia(data, classes, centers)
	delta := threshold * 100.0
	for i := 0; i < 100 && delta > threshold; i++ {
		classes = AssignToClass(data, centers)
		centers = ComputeCenters(data, k, classes)
		newi := ComputeInertia(data, classes, centers)
		delta = inertia - newi
		inertia = newi
		log.Debugf("Inertia: %f, delta: %f", inertia, delta)
	}

	return classes, centers, inertia, delta < threshold
}

// KmeansBestRepresentative finds the best representative among the data point of each cluster in parallel.
//
// It takes a matrix of data points and a matrix of centers as input.
// The best representative is the data point that is closest to the center of the cluster.
// Returns an array of integers containing the index of the best representative for each cluster.
func KmeansBestRepresentative(data *obiutils.Matrix[float64], centers *obiutils.Matrix[float64]) []int {
	bestRepresentative := make([]int, len(*centers))

	var wg sync.WaitGroup
	wg.Add(len(*centers))

	for j, center := range *centers {
		go func(j int, center []float64) {
			defer wg.Done()

			bestDistToCenter := math.MaxFloat64
			best := -1

			for i, row := range *data {
				dist := 0.0
				for d, val := range row {
					diff := val - center[d]
					dist += diff * diff
				}
				if dist < bestDistToCenter {
					bestDistToCenter = dist
					best = i
				}
			}

			if best == -1 {
				log.Fatalf("No representative found for cluster %d", j)
			}

			bestRepresentative[j] = best
		}(j, center)
	}

	wg.Wait()

	return bestRepresentative
}
