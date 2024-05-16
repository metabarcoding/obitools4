package obiutils

import (
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"
)

// InterfaceToString converts an interface value to a string.
//
// The function takes an interface{} value as a parameter and returns a string representation of that value.
// It returns the string representation and an error if any occurred during the conversion process.
func InterfaceToString(i interface{}) (val string, err error) {
	err = nil
	val = fmt.Sprintf("%v", i)
	return
}

// CastableToInt checks if the given input can be casted to an integer.
//
// i: the value to check for castability.
// bool: true if the value can be casted to an integer, false otherwise.
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

// InterfaceToBool converts an interface value to a boolean.
//
// It takes an interface{} as a parameter and returns a boolean value and an error.
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

// MapToMapInterface converts a map to a map of type map[string]interface{}.
//
// It takes an interface{} parameter `m` which represents the map to be converted.
//
// It returns a map[string]interface{} which is the converted map. If the input map is not of type map[string]interface{},
// it panics and logs an error message.
func MapToMapInterface(m interface{}) map[string]interface{} {
	if IsAMap(m) {
		reflectMap := reflect.ValueOf(m)
		keys := reflectMap.MapKeys()
		val := make(map[string]interface{}, len(keys))
		for k := range keys {
			val[keys[k].String()] = reflectMap.MapIndex(keys[k]).Interface()
		}

		return val
	}

	log.Panic("Invalid map type")
	return make(map[string]interface{})
}
