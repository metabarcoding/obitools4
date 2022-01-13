package goutils

import (
	"bytes"
	"encoding/gob"
)

type NotAnInteger struct {
	message string
}

func (m *NotAnInteger) Error() string {
	return m.message
}

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
		err = &NotAnInteger{"count attribute is not an integer"}
	}
	return
}

func CopyMap(dest, src map[string]interface{}) {
	buf := new(bytes.Buffer)
	gob.NewEncoder(buf).Encode(src)
	gob.NewDecoder(buf).Decode(&dest)
}
