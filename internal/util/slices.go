package util

func UniqueItems[T comparable](slices ...[]T) []T {
	asMap := make(map[T]any)
	unique := make([]T, 0)

	for _, slice := range slices {
		for _, item := range slice {
			asMap[item] = struct{}{}
		}
	}

	for item := range asMap {
		unique = append(unique, item)
	}

	return unique
}
