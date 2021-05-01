package utils

func StringsToSet(slice []string) map[string]bool {
	set := make(map[string]bool, len(slice))
	for _, e := range slice {
		set[e] = true
	}
	return set
}

func SetToStrings(set map[string]bool) []string {
	slice := make([]string, len(set))
	for k := range set {
		slice = append(slice, k)
	}
	return slice
}
