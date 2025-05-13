package obiutils

import (
	"errors"
	"fmt"
	"reflect"

	"golang.org/x/exp/constraints"
)

func MinMax[T constraints.Ordered](x, y T) (T, T) {
	if x < y {
		return x, y
	}
	return y, x
}

func MinMaxSlice[T constraints.Ordered](vec []T) (min, max T) {
	if len(vec) == 0 {
		panic("empty slice")
	}

	min = vec[0]
	max = vec[0]
	for _, v := range vec {
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}

	return
}

func MaxMap[K comparable, T constraints.Ordered](values map[K]T) (K, T, error) {
	var maxKey K
	var maxValue T

	if len(values) == 0 {
		return maxKey, maxValue, errors.New("Empty map")
	}

	first := true
	for k, v := range values {
		if v > maxValue || first {
			maxValue = v
			maxKey = k
			first = false
		}
	}

	return maxKey, maxValue, nil
}

func MinMap[K comparable, T constraints.Ordered](values map[K]T) (K, T, error) {
	var minKey K
	var minValue T
	if len(values) == 0 {
		return minKey, minValue, errors.New("Empty map")
	}

	first := true
	for k, v := range values {
		if v < minValue || first {
			minValue = v
			minKey = k
			first = false
		}
	}

	return minKey, minValue, nil
}

// Min returns the smallest element in a slice/array or map,
// or the value itself if data is a single comparable value.
// Returns an error if the container is empty or the type is unsupported.
func Min(data interface{}) (interface{}, error) {
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return nil, errors.New("empty slice or array")
		}
		return minFromIterable(v)
	case reflect.Map:
		if v.Len() == 0 {
			return nil, errors.New("empty map")
		}
		return minFromMap(v)
	default:
		if !isOrderedKind(v.Kind()) {
			return nil, fmt.Errorf("unsupported type: %s", v.Kind())
		}
		// single comparable value → return it
		return data, nil
	}
}

// Max returns the largest element in a slice/array or map,
// or the value itself if data is a single comparable value.
// Returns an error if the container is empty or the type is unsupported.
func Max(data interface{}) (interface{}, error) {
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return nil, errors.New("empty slice or array")
		}
		return maxFromIterable(v)
	case reflect.Map:
		if v.Len() == 0 {
			return nil, errors.New("empty map")
		}
		return maxFromMap(v)
	default:
		if !isOrderedKind(v.Kind()) {
			return nil, fmt.Errorf("unsupported type: %s", v.Kind())
		}
		return data, nil
	}
}

// maxFromIterable scans a slice/array to find the maximum.
func maxFromIterable(v reflect.Value) (interface{}, error) {
	var best reflect.Value
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if !isOrderedKind(elem.Kind()) {
			return nil, fmt.Errorf("unsupported element type: %s", elem.Kind())
		}
		if i == 0 || greater(elem, best) {
			best = elem
		}
	}
	return best.Interface(), nil
}

// minFromIterable finds min in slice or array.
func minFromIterable(v reflect.Value) (interface{}, error) {
	var minVal reflect.Value
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if !isOrderedKind(elem.Kind()) {
			return nil, fmt.Errorf("unsupported element type: %s", elem.Kind())
		}
		if i == 0 || less(elem, minVal) {
			minVal = elem
		}
	}
	return minVal.Interface(), nil
}

// maxFromMap scans map values to find the maximum.
func maxFromMap(v reflect.Value) (interface{}, error) {
	var best reflect.Value
	first := true
	for _, key := range v.MapKeys() {
		elem := v.MapIndex(key)
		if !isOrderedKind(elem.Kind()) {
			return nil, fmt.Errorf("unsupported element type: %s", elem.Kind())
		}
		if first || greater(elem, best) {
			best = elem
			first = false
		}
	}
	return best.Interface(), nil
}

// minFromMap finds min among map values.
func minFromMap(v reflect.Value) (interface{}, error) {
	var minVal reflect.Value
	first := true
	for _, key := range v.MapKeys() {
		elem := v.MapIndex(key)
		if !isOrderedKind(elem.Kind()) {
			return nil, fmt.Errorf("unsupported element type: %s", elem.Kind())
		}
		if first || less(elem, minVal) {
			minVal = elem
			first = false
		}
	}
	return minVal.Interface(), nil
}

// isOrderedKind reports whether k supports comparison ordering.
func isOrderedKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return true
	default:
		return false
	}
}

// less returns true if a < b for supported kinds.
func less(a, b reflect.Value) bool {
	switch a.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return a.Int() < b.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return a.Uint() < b.Uint()
	case reflect.Float32, reflect.Float64:
		return a.Float() < b.Float()
	case reflect.String:
		return a.String() < b.String()
	default:
		// should never happen if caller checked isOrderedKind
		return false
	}
}

// greater returns true if a > b for supported kinds.
func greater(a, b reflect.Value) bool {
	switch a.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return a.Int() > b.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return a.Uint() > b.Uint()
	case reflect.Float32, reflect.Float64:
		return a.Float() > b.Float()
	case reflect.String:
		return a.String() > b.String()
	default:
		return false
	}
}
