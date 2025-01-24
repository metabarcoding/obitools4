package obidefault

var _Quality_Shift_Input = byte(33)
var _Quality_Shift_Output = byte(33)
var _Read_Qualities = true

func SetReadQualitiesShift(shift byte) {
	_Quality_Shift_Input = shift
}

func ReadQualitiesShift() byte {
	return _Quality_Shift_Input
}

func SetWriteQualitiesShift(shift byte) {
	_Quality_Shift_Output = shift
}

func WriteQualitiesShift() byte {
	return _Quality_Shift_Output
}

func SetReadQualities(read bool) {
	_Read_Qualities = read
}

func ReadQualities() bool {
	return _Read_Qualities
}
