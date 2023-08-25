package obiutils

// Matrix is a generic type representing a matrix.
type Matrix[T any] [][]T

// Make2DArray generates a 2D array of type T with the specified number of rows and columns.
//
// Data is stored in a contiguous memory block in row-major order.
//
// Parameters:
// - rows: the number of rows in the 2D array.
// - cols: the number of columns in the 2D array.
//
// Return:
// - Matrix[T]: the generated 2D array of type T.
func Make2DArray[T any](rows, cols int) Matrix[T] {
	matrix := make(Matrix[T], rows)
	data := make([]T, cols*rows)
	for i := 0; i < rows; i++ {
		matrix[i] = data[i*cols : (i+1)*cols]
	}
	return matrix
}

// Row returns the i-th row of the matrix.
//
// Parameters:
//
//	i - the index of the row to retrieve.
//
// Return:
//
//	[]T - the i-th row of the matrix.
func (matrix *Matrix[T]) Column(i int) []T {
	r := make([]T, len(*matrix))
	for j := 0; j < len(*matrix); j++ {
		r[j] = (*matrix)[j][i]
	}
	return r
}
