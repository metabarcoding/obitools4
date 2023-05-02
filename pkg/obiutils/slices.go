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

func Reverse[S ~[]E, E any](s S, inplace bool) S {
	if !inplace {
		c := make([]E,len(s))
		copy(c,s)
		s = c
	}
    for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
        s[i], s[j] = s[j], s[i]
    }

	return s
}
