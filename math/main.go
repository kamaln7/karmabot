package math

// Min returns the smallest number of the two inputs
func Min(a, b int) int {
	if a < b {
		return a
	}

	return b
}
