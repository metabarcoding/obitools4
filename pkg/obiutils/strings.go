package obiutils

import "unsafe"

func UnsafeStringFreomBytes(data []byte) string {
	if len(data) > 0 {
		s := unsafe.String(unsafe.SliceData(data), len(data))
		return s
	}

	return ""
}
