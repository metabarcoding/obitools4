package obistats

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiutils"
	log "github.com/sirupsen/logrus"
	"math"
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
	for i, rowData := range *data {
		minDist := math.MaxFloat64
		for j, centerData := range *centers {
			dist := 0.0
			for d, val := range rowData {
				dist += math.Pow(val-centerData[d], 2)
			}
			if dist < minDist {
				minDist = dist
				classes[i] = j
			}
		}
	}
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
func ComputeCenters(data *obiutils.Matrix[float64], k int, classes []int) *obiutils.Matrix[float64] {
	centers := obiutils.Make2DArray[float64](k, len((*data)[0]))
	centers.Init(0.0)
	ns := make([]int, k)

	for i := range ns {
		ns[i] = 0
	}

	for i, row := range *data {
		ns[classes[i]]++
		for j, val := range row {
			centers[classes[i]][j] += val
		}
	}

	for i := range centers {
		for j := range centers[i] {
			centers[i][j] /= float64(ns[i])
		}
	}

	return &centers
}

// ComputeInertia computes the inertia of the given data and centers.
//
// Parameters:
// - data: A pointer to a Matrix of float64 representing the data.
// - centers: A pointer to a Matrix of float64 representing the centers.
//
// Return type:
// - float64: The computed inertia.
func ComputeInertia(data *obiutils.Matrix[float64], classes []int, centers *obiutils.Matrix[float64]) float64 {
	inertia := 0.0
	for i, row := range *data {
		for j, val := range row {
			inertia += math.Pow(val-(*centers)[classes[i]][j], 2)
		}
	}
	return inertia
}

// Kmeans performs the k-means clustering algorithm on the given data.
//
// if centers and *center is not nil, centers is considered as initialized
// and the number of classes (k) is set to the number of rows in centers.
// overwise, the number of classes is defined by the value of k.
//
// Parameters:
// - data: A pointer to a matrix containing the input data.
// - k: An integer representing the number of clusters.
// - centers: A pointer to a matrix representing the initial cluster centers.
//
// Returns:
// - A slice of integers representing the assigned class labels for each data point.
// - A pointer to a matrix representing the final cluster centers.

func Kmeans(data *obiutils.Matrix[float64],
	k int,
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
	threshold float64,
	centers *obiutils.Matrix[float64]) ([]int, *obiutils.Matrix[float64], float64, bool) {
	if centers == nil || *centers == nil {
		*centers = obiutils.Make2DArray[float64](k, len((*data)[0]))
		center_ids := SampleIntWithoutReplacemant(k, len(*data))
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

// KmeansBestRepresentative finds the best representative among the data point of each cluster.
//
// It takes a matrix of data points and a matrix of centers as input.
// The best representative is the data point that is closest to the center of the cluster.
// Returns an array of integers containing the index of the best representative for each cluster.
func KmeansBestRepresentative(data *obiutils.Matrix[float64], centers *obiutils.Matrix[float64]) []int {
	best_dist_to_centers := make([]float64, len(*centers))
	best_representative := make([]int, len(*centers))

	for i := range best_dist_to_centers {
		best_dist_to_centers[i] = math.MaxFloat64
	}

	for i, row := range *data {
		for j, center := range *centers {
			dist := 0.0
			for d, val := range row {
				dist += math.Pow(val-center[d], 2)
			}
			if dist < best_dist_to_centers[j] {
				best_dist_to_centers[j] = dist
				best_representative[j] = i
			}
		}
	}

	return best_representative
}
