package utilities

// Add a helper in utilities or here if utilities is not suitable
// utilities.MapKeys helps get keys from a map for error message
// (Assuming utilities package has a helper like this)
func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
