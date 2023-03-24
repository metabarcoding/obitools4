package obiutils

func Contains[T comparable](arr []T, x T) bool {
	for _, v := range arr {
		if v == x {
			return true
		}
	}
	return false
}

func LookFor[T comparable](arr []T, x T) int {
	for i, v := range arr {
		if v == x {
			return i
		}
	}
	return -1
}

func RemoveIndex[T comparable](s []T, index int) []T {
	return append(s[:index], s[index+1:]...)
}
