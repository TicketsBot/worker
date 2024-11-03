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

func ContainsFunc[T any](slice []T, f func(T) bool) bool {
	for _, el := range slice {
		if f(el) {
			return true
		}
	}

	return false
}

func HasIntersection[T comparable](slice []T, slice2 []T) bool {
	for _, el := range slice {
		for _, el2 := range slice2 {
			if el == el2 {
				return true
			}
		}
	}

	return false
}

func FindIntersection[T comparable](slice []T, slice2 []T) []T {
	var intersection []T
	for _, el := range slice {
		for _, el2 := range slice2 {
			if el == el2 {
				intersection = append(intersection, el)
			}
		}
	}

	return intersection
}

func Keys[T comparable, U any](m map[T]U) []T {
	keys := make([]T, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}
