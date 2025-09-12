package utils

import (
	"strings"
)

func SubSlice[T any](slice []T, from, to uint) []T {
	if int(from) > len(slice) {
		from = 0
	}
	if int(to) > len(slice) {
		to = uint(len(slice))
	}
	return slice[from:int(to)]
}

func CloneStringMap(m map[string]string, keysToLower bool) map[string]string {
	m2 := make(map[string]string)
	for k, v := range m {
		if keysToLower {
			k = strings.ToLower(k)
		}
		m2[k] = v
	}
	return m2
}
