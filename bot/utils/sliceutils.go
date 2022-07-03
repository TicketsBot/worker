package utils

func ToSlice[T any](slice ...T) []T {
	return slice
}

func Contains[T comparable](slice []T, target T) bool {
	for _, el := range slice {
		if el == target {
			return true
		}
	}

	return false
}
