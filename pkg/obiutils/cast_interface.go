package obiutils

import (
	"fmt"
	"reflect"
	"strconv"

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
	case bool:
		val = t
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
	case string:
		rep, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			err = &NotAnFloat64{"value attribute cannot be casted to an int value"}
		}
		val = int(rep)
	default:
		err = &NotAnInteger{"value attribute cannot be casted to an integer"}
	}
	return
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
	case string:
		val, err = strconv.ParseFloat(t, 10)
		if err != nil {
			err = &NotAnFloat64{"value attribute cannot be casted to a float value"}
		}
	default:
		err = &NotAnFloat64{"value attribute cannot be casted to a float value"}
	}
	return
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

func InterfaceToStringMap(i interface{}) (val map[string]string, err error) {
	err = nil

	switch i := i.(type) {
	case map[string]string:
		val = i
	case map[string]interface{}:
		val = make(map[string]string, len(i))
		for k, v := range i {
			val[k], err = InterfaceToString(v)
			if err != nil {
				return
			}
		}
	default:
		err = &NotAMapInt{"value attribute cannot be casted to a map[string]int"}
	}

	return
}

func InterfaceToStringSlice(i interface{}) (val []string, err error) {
	err = nil

	switch i := i.(type) {
	case []string:
		val = i
	case []interface{}:
		val = make([]string, len(i))
		for k, v := range i {
			val[k], err = InterfaceToString(v)
			if err != nil {
				return
			}
		}
	default:
		err = &NotAMapInt{"value attribute cannot be casted to a []string"}
	}

	return
}
