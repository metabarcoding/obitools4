package goutils

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"io"
	"os"
	"sync"
)

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
		err = &NotABoolean{"value attribute cannot be casted to an integer"}
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

// "CopyMap copies the contents of a map[string]interface{} to another map[string]interface{}."
//
// The function uses the gob package to encode the source map into a buffer and then decode the buffer
// into the destination map
func CopyMap(dest, src map[string]interface{}) {
	buf := new(bytes.Buffer)
	gob.NewEncoder(buf).Encode(src)
	gob.NewDecoder(buf).Decode(&dest)
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
