package obiutils

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