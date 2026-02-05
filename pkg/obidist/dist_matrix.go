package obidist

import (
	"fmt"
)

// DistMatrix represents a symmetric matrix stored as a triangular matrix.
// The diagonal has a constant value (typically 0 for distances, 1 for similarities).
// Only the upper triangle (i < j) is stored to save memory.
//
// For an n×n matrix, we store n(n-1)/2 values.
type DistMatrix struct {
	n            int       // Number of elements (matrix dimension)
	data         []float64 // Triangular storage: upper triangle only
	labels       []string  // Optional labels for rows/columns
	diagonalValue float64  // Value on the diagonal
}

// NewDistMatrix creates a new distance matrix of size n×n.
// All distances are initialized to 0.0, diagonal is 0.0.
func NewDistMatrix(n int) *DistMatrix {
	if n < 0 {
		panic("matrix size must be non-negative")
	}

	// Number of elements in upper triangle: n(n-1)/2
	size := n * (n - 1) / 2

	return &DistMatrix{
		n:             n,
		data:          make([]float64, size),
		labels:        make([]string, n),
		diagonalValue: 0.0,
	}
}

// NewDistMatrixWithLabels creates a new distance matrix with labels.
// Diagonal is 0.0 by default.
func NewDistMatrixWithLabels(labels []string) *DistMatrix {
	dm := NewDistMatrix(len(labels))
	copy(dm.labels, labels)
	return dm
}

// NewSimilarityMatrix creates a new similarity matrix of size n×n.
// All off-diagonal values are initialized to 0.0, diagonal is 1.0.
func NewSimilarityMatrix(n int) *DistMatrix {
	if n < 0 {
		panic("matrix size must be non-negative")
	}

	// Number of elements in upper triangle: n(n-1)/2
	size := n * (n - 1) / 2

	return &DistMatrix{
		n:             n,
		data:          make([]float64, size),
		labels:        make([]string, n),
		diagonalValue: 1.0,
	}
}

// NewSimilarityMatrixWithLabels creates a new similarity matrix with labels.
// Diagonal is 1.0.
func NewSimilarityMatrixWithLabels(labels []string) *DistMatrix {
	dm := NewSimilarityMatrix(len(labels))
	copy(dm.labels, labels)
	return dm
}

// Size returns the dimension of the matrix (n for an n×n matrix).
func (dm *DistMatrix) Size() int {
	return dm.n
}

// indexFor computes the index in the data array for element (i, j).
// Assumes i < j (upper triangle).
//
// The upper triangle is stored row by row:
// (0,1), (0,2), ..., (0,n-1), (1,2), (1,3), ..., (1,n-1), (2,3), ...
//
// For element (i, j) where i < j:
// index = i*(n-1) + j - 1 - i*(i+1)/2
//
// This can be simplified to:
// index = i*n - i*(i+1)/2 + j - i - 1
//       = i*(n - (i+1)/2 - 1) + j - 1
//       = i*(n - 1 - i/2 - 1/2) + j - 1
//
// But the clearest formula is:
// index = i*n - i*(i+3)/2 + j - 1
func (dm *DistMatrix) indexFor(i, j int) int {
	if i >= j {
		panic(fmt.Sprintf("indexFor expects i < j, got i=%d, j=%d", i, j))
	}
	// Formula: number of elements in previous rows + position in current row
	// Previous rows (0 to i-1): sum from k=0 to i-1 of (n-1-k) = i*n - i*(i+1)/2
	// Current row position: j - i - 1
	return i*dm.n - i*(i+1)/2 + j - i - 1
}

// Get returns the value at position (i, j).
// The matrix is symmetric, so Get(i, j) == Get(j, i).
// The diagonal returns the diagonalValue (0.0 for distances, 1.0 for similarities).
func (dm *DistMatrix) Get(i, j int) float64 {
	if i < 0 || i >= dm.n || j < 0 || j >= dm.n {
		panic(fmt.Sprintf("indices out of bounds: i=%d, j=%d, n=%d", i, j, dm.n))
	}

	// Diagonal: return the diagonal value
	if i == j {
		return dm.diagonalValue
	}

	// Ensure i < j for indexing
	if i > j {
		i, j = j, i
	}

	return dm.data[dm.indexFor(i, j)]
}

// Set sets the value at position (i, j).
// The matrix is symmetric, so Set(i, j, v) also sets (j, i) to v.
// Setting the diagonal (i == j) is ignored (diagonal has a fixed value).
func (dm *DistMatrix) Set(i, j int, value float64) {
	if i < 0 || i >= dm.n || j < 0 || j >= dm.n {
		panic(fmt.Sprintf("indices out of bounds: i=%d, j=%d, n=%d", i, j, dm.n))
	}

	// Ignore diagonal assignments (diagonal has a fixed value)
	if i == j {
		return
	}

	// Ensure i < j for indexing
	if i > j {
		i, j = j, i
	}

	dm.data[dm.indexFor(i, j)] = value
}

// GetLabel returns the label for element i.
func (dm *DistMatrix) GetLabel(i int) string {
	if i < 0 || i >= dm.n {
		panic(fmt.Sprintf("index out of bounds: i=%d, n=%d", i, dm.n))
	}
	return dm.labels[i]
}

// SetLabel sets the label for element i.
func (dm *DistMatrix) SetLabel(i int, label string) {
	if i < 0 || i >= dm.n {
		panic(fmt.Sprintf("index out of bounds: i=%d, n=%d", i, dm.n))
	}
	dm.labels[i] = label
}

// Labels returns a copy of all labels.
func (dm *DistMatrix) Labels() []string {
	labels := make([]string, dm.n)
	copy(labels, dm.labels)
	return labels
}

// GetRow returns the i-th row of the distance matrix.
// The returned slice is a copy.
func (dm *DistMatrix) GetRow(i int) []float64 {
	if i < 0 || i >= dm.n {
		panic(fmt.Sprintf("index out of bounds: i=%d, n=%d", i, dm.n))
	}

	row := make([]float64, dm.n)
	for j := 0; j < dm.n; j++ {
		row[j] = dm.Get(i, j)
	}
	return row
}

// GetColumn returns the j-th column of the distance matrix.
// Since the matrix is symmetric, GetColumn(j) == GetRow(j).
// The returned slice is a copy.
func (dm *DistMatrix) GetColumn(j int) []float64 {
	return dm.GetRow(j)
}

// MinDistance returns the minimum non-zero distance in the matrix,
// along with the indices (i, j) where it occurs.
// Returns (0.0, -1, -1) if the matrix is empty or all distances are 0.
func (dm *DistMatrix) MinDistance() (float64, int, int) {
	if dm.n <= 1 {
		return 0.0, -1, -1
	}

	minDist := -1.0
	minI, minJ := -1, -1

	for i := 0; i < dm.n-1; i++ {
		for j := i + 1; j < dm.n; j++ {
			dist := dm.Get(i, j)
			if minDist < 0 || dist < minDist {
				minDist = dist
				minI = i
				minJ = j
			}
		}
	}

	if minI < 0 {
		return 0.0, -1, -1
	}

	return minDist, minI, minJ
}

// MaxDistance returns the maximum distance in the matrix,
// along with the indices (i, j) where it occurs.
// Returns (0.0, -1, -1) if the matrix is empty.
func (dm *DistMatrix) MaxDistance() (float64, int, int) {
	if dm.n <= 1 {
		return 0.0, -1, -1
	}

	maxDist := -1.0
	maxI, maxJ := -1, -1

	for i := 0; i < dm.n-1; i++ {
		for j := i + 1; j < dm.n; j++ {
			dist := dm.Get(i, j)
			if maxDist < 0 || dist > maxDist {
				maxDist = dist
				maxI = i
				maxJ = j
			}
		}
	}

	if maxI < 0 {
		return 0.0, -1, -1
	}

	return maxDist, maxI, maxJ
}

// Copy creates a deep copy of the matrix.
func (dm *DistMatrix) Copy() *DistMatrix {
	newDM := &DistMatrix{
		n:             dm.n,
		data:          make([]float64, len(dm.data)),
		labels:        make([]string, dm.n),
		diagonalValue: dm.diagonalValue,
	}

	copy(newDM.data, dm.data)
	copy(newDM.labels, dm.labels)

	return newDM
}

// ToFullMatrix returns a full n×n matrix representation.
// This allocates n² values, so use only when needed.
func (dm *DistMatrix) ToFullMatrix() [][]float64 {
	matrix := make([][]float64, dm.n)
	for i := 0; i < dm.n; i++ {
		matrix[i] = make([]float64, dm.n)
		for j := 0; j < dm.n; j++ {
			matrix[i][j] = dm.Get(i, j)
		}
	}
	return matrix
}
