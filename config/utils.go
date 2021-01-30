package config

import "strings"

func cloneStringMap(m map[string]string, keysToLower bool) map[string]string {
	m2 := make(map[string]string)
	for k, v := range m {
		if keysToLower {
			k = strings.ToLower(k)
		}
		m2[k] = v
	}
	return m2
}
