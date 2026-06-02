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

func FilterMinSlice[T constraints.Ordered](vec []T, minimum T) []T {
	result := make([]T, 0, len(vec))
	for _, v := range vec {
		if v >= minimum {
			result = append(result, v)
		}
	}
	return result
}

func FilterMaxSlice[T constraints.Ordered](vec []T, maximum T) []T {
	result := make([]T, 0, len(vec))
	for _, v := range vec {
		if v <= maximum {
			result = append(result, v)
		}
	}
	return result
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

func FilterMinMap[K comparable, T constraints.Ordered](values map[K]T, minimum T) map[K]T {
	result := make(map[K]T)
	for k, v := range values {
		if v >= minimum {
			result[k] = v
		}
	}
	return result
}

func FilterMaxMap[K comparable, T constraints.Ordered](values map[K]T, maximum T) map[K]T {
	result := make(map[K]T)
	for k, v := range values {
		if v <= maximum {
			result[k] = v
		}
	}
	return result
}

func SaturatingSubSlice[T Numeric](vec []T, sub T) []T {
	result := make([]T, len(vec))
	for i, v := range vec {
		if v > sub {
			result[i] = v - sub
		}
	}
	return result
}

func SaturatingSubMap[K comparable, T Numeric](values map[K]T, sub T) map[K]T {
	result := make(map[K]T)
	for k, v := range values {
		if v > sub {
			result[k] = v - sub
		}
	}
	return result
}

// Min returns the smallest element in a slice/array or map,
// or the value itself if data is a single comparable value.
// Returns an error if the container is empty or the type is unsupported.
func Min(data interface{}) (interface{}, error) {
	v := reflect.ValueOf(data)
	method := v.MethodByName("Min")
	if method.IsValid() {
		result := method.Call(nil)[0].Interface()
		return result, nil
	}

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

	method := v.MethodByName("Max")
	if method.IsValid() {
		result := method.Call(nil)[0].Interface()
		return result, nil
	}

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

func FilterMin(data interface{}, minimum interface{}) (interface{}, error) {
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return nil, errors.New("empty slice or array")
		}
		return filterMinFromIterable(v, minimum)
	case reflect.Map:
		if v.Len() == 0 {
			return nil, errors.New("empty map")
		}
		return filterMinFromMap(v, minimum)
	default:
		if !isOrderedKind(v.Kind()) {
			return nil, fmt.Errorf("unsupported type: %s", v.Kind())
		}
		return data, nil
	}
}

func FilterMax(data interface{}, maximum interface{}) (interface{}, error) {
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return nil, errors.New("empty slice or array")
		}
		return filterMaxFromIterable(v, maximum)
	case reflect.Map:
		if v.Len() == 0 {
			return nil, errors.New("empty map")
		}
		return filterMaxFromMap(v, maximum)
	default:
		if !isOrderedKind(v.Kind()) {
			return nil, fmt.Errorf("unsupported type: %s", v.Kind())
		}
		return data, nil
	}
}

func SaturatingSub(data interface{}, sub interface{}) (interface{}, error) {
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return saturatingSubFromIterable(v, sub)
	case reflect.Map:
		return saturatingSubFromMap(v, sub)
	default:
		if !isNumericKind(v.Kind()) {
			return nil, fmt.Errorf("unsupported type: %s", v.Kind())
		}
		r, err := saturatingSubValues(v, reflect.ValueOf(sub))
		if err != nil {
			return nil, err
		}
		return r.Interface(), nil
	}
}

func saturatingSubFromIterable(v reflect.Value, sub interface{}) (interface{}, error) {
	subVal := reflect.ValueOf(sub)
	result := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
	for i := 0; i < v.Len(); i++ {
		r, err := saturatingSubValues(v.Index(i), subVal)
		if err != nil {
			return nil, err
		}
		result.Index(i).Set(r)
	}
	return result.Interface(), nil
}

func saturatingSubFromMap(v reflect.Value, sub interface{}) (interface{}, error) {
	subVal := reflect.ValueOf(sub)
	result := reflect.MakeMap(v.Type())
	for _, key := range v.MapKeys() {
		r, err := saturatingSubValues(v.MapIndex(key), subVal)
		if err != nil {
			return nil, err
		}
		if !r.IsZero() {
			result.SetMapIndex(key, r)
		}
	}
	return result.Interface(), nil
}

func saturatingSubValues(a, b reflect.Value) (reflect.Value, error) {
	result := reflect.New(a.Type()).Elem()
	switch a.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if av, bv := a.Int(), b.Int(); av > bv {
			result.SetInt(av - bv)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if av, bv := a.Uint(), b.Uint(); av > bv {
			result.SetUint(av - bv)
		}
	case reflect.Float32, reflect.Float64:
		if av, bv := a.Float(), b.Float(); av > bv {
			result.SetFloat(av - bv)
		}
	default:
		return reflect.Value{}, fmt.Errorf("unsupported type for saturating subtraction: %s", a.Kind())
	}
	return result, nil
}

// maxFromIterable scans a slice/array to find the maximum.
func maxFromIterable(v reflect.Value) (interface{}, error) {
	var best reflect.Value
	for i := 0; i < v.Len(); i++ {
		elem := unwrapInterface(v.Index(i))
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
		elem := unwrapInterface(v.Index(i))
		if !isOrderedKind(elem.Kind()) {
			return nil, fmt.Errorf("unsupported element type: %s", elem.Kind())
		}
		if i == 0 || less(elem, minVal) {
			minVal = elem
		}
	}
	return minVal.Interface(), nil
}

func filterMinFromIterable(v reflect.Value, minimum interface{}) (interface{}, error) {
	minVal := reflect.ValueOf(minimum)
	result := reflect.MakeSlice(v.Type(), 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		elem := unwrapInterface(v.Index(i))
		if !isOrderedKind(elem.Kind()) {
			return nil, fmt.Errorf("unsupported element type: %s", elem.Kind())
		}
		if !less(elem, minVal) { // elem >= minimum
			result = reflect.Append(result, elem)
		}
	}
	return result.Interface(), nil
}

func filterMaxFromIterable(v reflect.Value, maximum interface{}) (interface{}, error) {
	maxVal := reflect.ValueOf(maximum)
	result := reflect.MakeSlice(v.Type(), 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		elem := unwrapInterface(v.Index(i))
		if !isOrderedKind(elem.Kind()) {
			return nil, fmt.Errorf("unsupported element type: %s", elem.Kind())
		}
		if !greater(elem, maxVal) { // elem <= maximum
			result = reflect.Append(result, elem)
		}
	}
	return result.Interface(), nil
}

// whichMaxFromIterable returns the index of the maximum element in a slice/array.
func whichMaxFromIterable(v reflect.Value) (int, error) {
	var best reflect.Value
	bestIdx := 0
	for i := 0; i < v.Len(); i++ {
		elem := unwrapInterface(v.Index(i))
		if !isOrderedKind(elem.Kind()) {
			return 0, fmt.Errorf("unsupported element type: %s", elem.Kind())
		}
		if i == 0 || greater(elem, best) {
			best = elem
			bestIdx = i
		}
	}
	return bestIdx, nil
}

// whichMinFromIterable returns the index of the minimum element in a slice/array.
func whichMinFromIterable(v reflect.Value) (int, error) {
	var minVal reflect.Value
	minIdx := 0
	for i := 0; i < v.Len(); i++ {
		elem := unwrapInterface(v.Index(i))
		if !isOrderedKind(elem.Kind()) {
			return 0, fmt.Errorf("unsupported element type: %s", elem.Kind())
		}
		if i == 0 || less(elem, minVal) {
			minVal = elem
			minIdx = i
		}
	}
	return minIdx, nil
}

// whichMaxFromMap returns the key associated with the maximum value in a map.
func whichMaxFromMap(v reflect.Value) (interface{}, error) {
	var best reflect.Value
	var bestKey reflect.Value
	first := true
	for _, key := range v.MapKeys() {
		elem := unwrapInterface(v.MapIndex(key))
		if !isOrderedKind(elem.Kind()) {
			return nil, fmt.Errorf("unsupported element type: %s", elem.Kind())
		}
		if first || greater(elem, best) {
			best = elem
			bestKey = key
			first = false
		}
	}
	return bestKey.Interface(), nil
}

// whichMinFromMap returns the key associated with the minimum value in a map.
func whichMinFromMap(v reflect.Value) (interface{}, error) {
	var minVal reflect.Value
	var minKey reflect.Value
	first := true
	for _, key := range v.MapKeys() {
		elem := unwrapInterface(v.MapIndex(key))
		if !isOrderedKind(elem.Kind()) {
			return nil, fmt.Errorf("unsupported element type: %s", elem.Kind())
		}
		if first || less(elem, minVal) {
			minVal = elem
			minKey = key
			first = false
		}
	}
	return minKey.Interface(), nil
}

// WhichMax returns the key (for a map) or index (for a slice/array) of the maximum value.
func WhichMax(data interface{}) (interface{}, error) {
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return nil, errors.New("empty slice or array")
		}
		return whichMaxFromIterable(v)
	case reflect.Map:
		if v.Len() == 0 {
			return nil, errors.New("empty map")
		}
		return whichMaxFromMap(v)
	default:
		return nil, fmt.Errorf("unsupported type: %s", v.Kind())
	}
}

// WhichMin returns the key (for a map) or index (for a slice/array) of the minimum value.
func WhichMin(data interface{}) (interface{}, error) {
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return nil, errors.New("empty slice or array")
		}
		return whichMinFromIterable(v)
	case reflect.Map:
		if v.Len() == 0 {
			return nil, errors.New("empty map")
		}
		return whichMinFromMap(v)
	default:
		return nil, fmt.Errorf("unsupported type: %s", v.Kind())
	}
}

func filterMinFromMap(v reflect.Value, minimum interface{}) (interface{}, error) {
	minVal := reflect.ValueOf(minimum)
	result := reflect.MakeMap(v.Type())
	for _, key := range v.MapKeys() {
		elem := unwrapInterface(v.MapIndex(key))
		if !isOrderedKind(elem.Kind()) {
			return nil, fmt.Errorf("unsupported element type: %s", elem.Kind())
		}
		if !less(elem, minVal) { // elem >= minimum
			result.SetMapIndex(key, elem)
		}
	}
	return result.Interface(), nil
}

func filterMaxFromMap(v reflect.Value, maximum interface{}) (interface{}, error) {
	maxVal := reflect.ValueOf(maximum)
	result := reflect.MakeMap(v.Type())
	for _, key := range v.MapKeys() {
		elem := unwrapInterface(v.MapIndex(key))
		if !isOrderedKind(elem.Kind()) {
			return nil, fmt.Errorf("unsupported element type: %s", elem.Kind())
		}
		if !greater(elem, maxVal) { // elem <= maximum
			result.SetMapIndex(key, elem)
		}
	}
	return result.Interface(), nil
}

// maxFromMap scans map values to find the maximum.
func maxFromMap(v reflect.Value) (interface{}, error) {
	var best reflect.Value
	first := true
	for _, key := range v.MapKeys() {
		elem := unwrapInterface(v.MapIndex(key))
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
		elem := unwrapInterface(v.MapIndex(key))
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

func isNumericKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// unwrapInterface returns v.Elem() when v holds an interface value, otherwise v unchanged.
// This is necessary when iterating map[string]interface{} or []interface{} via reflection:
// the element Kind is reflect.Interface, not the underlying concrete type.
func unwrapInterface(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Interface {
		return v.Elem()
	}
	return v
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
