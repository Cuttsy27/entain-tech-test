package utils

// Returns true if a given string is present in a slice of strings, else false.
func Contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
