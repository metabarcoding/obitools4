package obiutils

type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Float interface {
	~float32 | ~float64
}
type Numeric interface {
	Integer | Float
}

type Vector[T any] []T
type Matrix[T any] []Vector[T]

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

func Make2DNumericArray[T Numeric](rows, cols int, zeroed bool) Matrix[T] {
	matrix := make(Matrix[T], rows)
	data := make([]T, cols*rows)

	if zeroed {
		for i := range data {
			data[i] = 0
		}
	}

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

// Rows returns the specified rows of the matrix.
//
// The function takes one or more integer arguments representing the indices of the rows to be returned.
// It returns a new matrix containing the specified rows.
func (matrix *Matrix[T]) Rows(i ...int) Matrix[T] {
	res := make([]Vector[T], len(i))

	for j, idx := range i {
		res[j] = (*matrix)[idx]
	}
	return res
}

// Dim returns the dimensions of the Matrix.
//
// It takes no parameters.
// It returns two integers: the number of rows and the number of columns.
func (matrix *Matrix[T]) Dim() (int, int) {
	return len(*matrix), len((*matrix)[0])
}
