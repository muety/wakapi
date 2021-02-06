package models

type IntervalKey []string

func (k *IntervalKey) HasAlias(s string) bool {
	for _, e := range *k {
		if e == s {
			return true
		}
	}
	return false
}
