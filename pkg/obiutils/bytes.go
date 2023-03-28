package obiutils

func InPlaceToLower(data []byte) []byte {
	for i,l := range data {
		if l >= 'A' && l <='Z' {
			data[i]|=32
		}
	}

	return data
}