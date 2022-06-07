package utils

func Ptr[T any](t T) *T {
	return &t
}

func PtrElems[T any](t []T) []*T {
	arr := make([]*T, len(t))
	for i, v := range t {
		arr[i] = &v
	}

	return arr
}

func Slice[T any](v ...T) []T {
	return v
}
