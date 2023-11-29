package utils

func MergeMaps[K comparable, V any](MyMap1 map[K]V, MyMap2 map[K]V) map[K]V {
	merged := make(map[K]V)
	for key, val := range MyMap1 {
		merged[key] = val
	}
	for key, val := range MyMap2 {
		merged[key] = val
	}
	return merged
}
