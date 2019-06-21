package util

// Contains Returns true is the needle string
// is found in the haystack.
func Contains(needle string, haystack []string) bool {
	for _, element := range haystack {
		if needle == element {
			return true
		}
	}
	return false
}
