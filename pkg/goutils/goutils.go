package goutils

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

// It converts an interface{} to a bool, and returns an error if the interface{} cannot be converted
// to a bool
func InterfaceToBool(i interface{}) (val bool, err error) {

	err = nil
	val = false

	switch t := i.(type) {
	case int:
		val = t != 0
	case int8:
		val = t != 0 // standardizes across systems
	case int16:
		val = t != 0 // standardizes across systems
	case int32:
		val = t != 0 // standardizes across systems
	case int64:
		val = t != 0 // standardizes across systems
	case float32:
		val = t != 0 // standardizes across systems
	case float64:
		val = t != 0 // standardizes across systems
	case uint8:
		val = t != 0 // standardizes across systems
	case uint16:
		val = t != 0 // standardizes across systems
	case uint32:
		val = t != 0 // standardizes across systems
	case uint64:
		val = t != 0 // standardizes across systems
	default:
		err = &NotABoolean{"value attribute cannot be casted to a boolean"}
	}
	return
}

// If the interface{} can be cast to an int, return true.
func CastableToInt(i interface{}) bool {

	switch i.(type) {
	case int,
		int8, int16, int32, int64,
		float32, float64,
		uint8, uint16, uint32, uint64:
		return true
	default:
		return false
	}
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

// Read a whole file into the memory and store it as array of lines
// It reads a file line by line, and returns a slice of strings, one for each line
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

func Contains[T comparable](arr []T, x T) bool {
	for _, v := range arr {
		if v == x {
			return true
		}
	}
	return false
}

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

// Marshal is a UTF-8 friendly marshaler.  Go's json.Marshal is not UTF-8
// friendly because it replaces the valid UTF-8 and JSON characters "&". "<",
// ">" with the "slash u" unicode escaped forms (e.g. \u0026).  It preemptively
// escapes for HTML friendliness.  Where text may include any of these
// characters, json.Marshal should not be used. Playground of Go breaking a
// title: https://play.golang.org/p/o2hiX0c62oN
func JsonMarshal(i interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(i)
	return bytes.TrimRight(buffer.Bytes(), "\n"), err
}

func IsAMap(value interface{}) bool {
	return reflect.TypeOf(value).Kind() == reflect.Map
}

func IsAnArray(value interface{}) bool {
	return reflect.TypeOf(value).Kind() == reflect.Array
}

func IsASlice(value interface{}) bool {
	return reflect.TypeOf(value).Kind() == reflect.Slice
}

func HasLength(value interface{}) bool {
	_, ok := value.(interface{ Len() int })
	return IsAMap(value) || IsAnArray(value) || IsASlice(value) || ok
}
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
