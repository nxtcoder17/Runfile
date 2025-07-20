package functions

import "maps"

func MapMerge[K comparable, V any](items ...map[K]V) map[K]V {
	result := make(map[K]V)
	for i := range items {
		maps.Copy(result, items[i])
	}
	return result
}

func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
