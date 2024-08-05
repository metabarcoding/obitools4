package obiutils

import "unsafe"

func UnsafeStringFromBytes(data []byte) string {
	if len(data) > 0 {
		s := unsafe.String(unsafe.SliceData(data), len(data))
		return s
	}

	return ""
}
