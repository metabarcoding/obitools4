package obiutils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"sync"

	"github.com/barkimedes/go-deepcopy"
)

// InterfaceToInt converts a interface{} to an integer value if possible.
// If not a "NotAnInteger" error is returned via the err
// return value and val is set to 0.
func InterfaceToString(i interface{}) (val string, err error) {
	err = nil
	val = fmt.Sprintf("%v", i)
	return
}

// NotAnInteger defines a new type of Error : "NotAnInteger"
type NotAnInteger struct {
	message string
}

// Error() retreives the error message associated to the "NotAnInteger"
// error. Tha addition of that Error message make the "NotAnInteger"
// complying with the error interface
func (m *NotAnInteger) Error() string {
	return m.message
}

// InterfaceToInt converts a interface{} to an integer value if possible.
// If not a "NotAnInteger" error is returned via the err
// return value and val is set to 0.
func InterfaceToInt(i interface{}) (val int, err error) {

	err = nil
	val = 0

	switch t := i.(type) {
	case int:
		val = t
	case int8:
		val = int(t) // standardizes across systems
	case int16:
		val = int(t) // standardizes across systems
	case int32:
		val = int(t) // standardizes across systems
	case int64:
		val = int(t) // standardizes across systems
	case float32:
		val = int(t) // standardizes across systems
	case float64:
		val = int(t) // standardizes across systems
	case uint8:
		val = int(t) // standardizes across systems
	case uint16:
		val = int(t) // standardizes across systems
	case uint32:
		val = int(t) // standardizes across systems
	case uint64:
		val = int(t) // standardizes across systems
	default:
		err = &NotAnInteger{"value attribute cannot be casted to an integer"}
	}
	return
}

// NotAnInteger defines a new type of Error : "NotAnInteger"
type NotAnFloat64 struct {
	message string
}

// Error() retreives the error message associated to the "NotAnInteger"
// error. Tha addition of that Error message make the "NotAnInteger"
// complying with the error interface
func (m *NotAnFloat64) Error() string {
	return m.message
}

// InterfaceToInt converts a interface{} to an integer value if possible.
// If not a "NotAnInteger" error is returned via the err
// return value and val is set to 0.
func InterfaceToFloat64(i interface{}) (val float64, err error) {

	err = nil
	val = 0

	switch t := i.(type) {
	case int:
		val = float64(t)
	case int8:
		val = float64(t) // standardizes across systems
	case int16:
		val = float64(t) // standardizes across systems
	case int32:
		val = float64(t) // standardizes across systems
	case int64:
		val = float64(t) // standardizes across systems
	case float32:
		val = float64(t) // standardizes across systems
	case float64:
		val = t // standardizes across systems
	case uint8:
		val = float64(t) // standardizes across systems
	case uint16:
		val = float64(t) // standardizes across systems
	case uint32:
		val = float64(t) // standardizes across systems
	case uint64:
		val = float64(t) // standardizes across systems
	default:
		err = &NotAnFloat64{"value attribute cannot be casted to a float value"}
	}
	return
}

// NotABoolean defines a new type of Error : "NotAMapInt"
type NotAMapInt struct {
	message string
}

// Error() retreives the error message associated to the "NotAnInteger"
// error. Tha addition of that Error message make the "NotAnInteger"
// complying with the error interface
func (m *NotAMapInt) Error() string {
	return m.message
}

func InterfaceToIntMap(i interface{}) (val map[string]int, err error) {
	err = nil

	switch i := i.(type) {
	case map[string]int:
		val = i
	case map[string]interface{}:
		val = make(map[string]int, len(i))
		for k, v := range i {
			val[k], err = InterfaceToInt(v)
			if err != nil {
				return
			}
		}
	case map[string]float64:
		val = make(map[string]int, len(i))
		for k, v := range i {
			val[k] = int(v)
		}
	default:
		err = &NotAMapInt{"value attribute cannot be casted to a map[string]int"}
	}

	return
}

// NotABoolean defines a new type of Error : "NotAMapInt"
type NotAMapFloat64 struct {
	message string
}

// Error() retreives the error message associated to the "NotAnInteger"
// error. Tha addition of that Error message make the "NotAnInteger"
// complying with the error interface
func (m *NotAMapFloat64) Error() string {
	return m.message
}

func InterfaceToFloat64Map(i interface{}) (val map[string]float64, err error) {
	err = nil

	switch i := i.(type) {
	case map[string]float64:
		val = i
	case map[string]interface{}:
		val = make(map[string]float64, len(i))
		for k, v := range i {
			val[k], err = InterfaceToFloat64(v)
			if err != nil {
				return
			}
		}
	case map[string]int:
		val = make(map[string]float64, len(i))
		for k, v := range i {
			val[k] = float64(v)
		}
	default:
		err = &NotAMapFloat64{"value attribute cannot be casted to a map[string]float64"}
	}

	return
}

// InterfaceToFloat64Slice converts an interface{} to a []float64 slice.
//
// It takes an interface{} parameter and returns a slice of float64 values and an error.
func InterfaceToFloat64Slice(i interface{}) ([]float64, error) {
	switch i := i.(type) {
	case []float64:
		return i, nil
	case []interface{}:
		val := make([]float64, len(i))
		for k, v := range i {
			if x, err := InterfaceToFloat64(v); err != nil {
				return nil, err
			} else {
				val[k] = x
			}
		}
		return val, nil
	case []int:
		val := make([]float64, len(i))
		for k, v := range i {
			val[k] = float64(v)
		}
		return val, nil
	default:
		return nil, &NotAMapFloat64{"value attribute cannot be casted to a []float64"}
	}
}

// InterfaceToIntSlice converts an interface{} to a []int slice.
//
// It takes an interface{} parameter and returns a slice of int values and an error.
func InterfaceToIntSlice(i interface{}) ([]int, error) {

	switch i := i.(type) {
	case []int:
		return i, nil
	case []interface{}:
		val := make([]int, len(i))
		for k, v := range i {
			if x, err := InterfaceToInt(v); err != nil {
				return nil, err
			} else {
				val[k] = x
			}
		}
		return val, nil
	case []float64:
		val := make([]int, len(i))
		for k, v := range i {
			val[k] = int(v + 0.5)
		}
		return val, nil
	case Vector[float64]:
		val := make([]int, len(i))
		for k, v := range i {
			val[k] = int(v + 0.5)
		}
		return val, nil
	default:
		return nil, &NotAMapInt{"value attribute cannot be casted to a []int"}
	}
}

// NotABoolean defines a new type of Error : "NotABoolean"
type NotABoolean struct {
	message string
}

// Error() retreives the error message associated to the "NotABoolean"
// error. Tha addition of that Error message make the "NotABoolean"
// complying with the error interface
func (m *NotABoolean) Error() string {
	return m.message
}

// > It copies the contents of the `src` map into the `dest` map, but if the value is a map, slice, or
// array, it makes a deep copy of it
func MustFillMap(dest, src map[string]interface{}) {

	for k, v := range src {
		if IsAMap(v) || IsASlice(v) || IsAnArray(v) {
			v = deepcopy.MustAnything(v)
		}
		dest[k] = v
	}
}

// ReadLines reads the lines from a file specified by the given path.
//
// Read a whole file into the memory and store it as array of lines
// It reads a file line by line, and returns a slice of strings, one for each line
//
// It takes a single parameter:
// - path: a string representing the path of the file to read.
//
// It returns a slice of strings containing the lines read from the file, and an error if any occurred.
func ReadLines(path string) (lines []string, err error) {
	var (
		file   *os.File
		part   []byte
		prefix bool
	)
	if file, err = os.Open(path); err != nil {
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	buffer := bytes.NewBuffer(make([]byte, 0))
	for {
		if part, prefix, err = reader.ReadLine(); err != nil {
			break
		}
		buffer.Write(part)
		if !prefix {
			lines = append(lines, buffer.String())
			buffer.Reset()
		}
	}
	if err == io.EOF {
		err = nil
	}
	return
}

// AtomicCounter creates and returns a function that generates a unique integer value each time it is called.
//
// The function takes an optional initial value as a parameter. If an initial value is provided, the generated
// integers will start from that value. If no initial value is provided, the generated integers will start from 0.
//
// The function is thread safe.
//
// The function returns a closure that can be called to retrieve the next integer in the sequence.
func AtomicCounter(initial ...int) func() int {
	counterMutex := sync.Mutex{}
	counter := 0
	if len(initial) > 0 {
		counter = initial[0]
	}

	nextCounter := func() int {
		counterMutex.Lock()
		defer counterMutex.Unlock()
		val := counter
		counter++

		return val
	}

	return nextCounter
}

// JsonMarshal marshals an interface into JSON format.
//
// JsonMarshal is a UTF-8 friendly marshaler.  Go's json.Marshal is not UTF-8
// friendly because it replaces the valid UTF-8 and JSON characters "&". "<",
// ">" with the "slash u" unicode escaped forms (e.g. \u0026).  It preemptively
// escapes for HTML friendliness.  Where text may include any of these
// characters, json.Marshal should not be used. Playground of Go breaking a
// title: https://play.golang.org/p/o2hiX0c62oN
//
// It takes an interface as a parameter and returns a byte slice and an error.
func JsonMarshal(i interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(i)
	return bytes.TrimRight(buffer.Bytes(), "\n"), err
}

// IsAMap checks if the given value is a map.
//
// value: the value to be checked.
// returns: a boolean indicating if the value is a map.
func IsAMap(value interface{}) bool {
	return reflect.TypeOf(value).Kind() == reflect.Map
}

// IsAnArray checks if the given value is an array.
//
// value: The value to be checked.
// Returns: true if the value is an array, false otherwise.
func IsAnArray(value interface{}) bool {
	return reflect.TypeOf(value).Kind() == reflect.Array
}

// IsASlice determines if the given value is a slice.
//
// value: the value to check.
// bool: true if the value is a slice, false otherwise.
func IsASlice(value interface{}) bool {
	return reflect.TypeOf(value).Kind() == reflect.Slice
}

// HasLength checks if the given value has a length.
//
// value: The value to be checked.
// bool: Returns true if the value has a length, false otherwise.
func HasLength(value interface{}) bool {
	_, ok := value.(interface{ Len() int })
	return IsAMap(value) || IsAnArray(value) || IsASlice(value) || ok
}

// Len returns the length of the given value.
//
// It accepts a single parameter:
// - value: an interface{} that represents the value whose length is to be determined.
//
// It returns an int, which represents the length of the value.
func Len(value interface{}) int {
	l := 1

	if IsAMap(value) || IsAnArray(value) || IsASlice(value) {
		vc := reflect.ValueOf(value)
		l = vc.Len()
	}

	if vc, ok := value.(interface{ Len() int }); ok {
		l = vc.Len()
	}

	return l
}
