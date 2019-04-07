package goskiplist

//Min min of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

//Max max of two ints
func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

// TurnEven if int is even return it,
// else return next even int
func turnEven(a int) int {
	if a&1 == 1 {
		return a + 1
	}
	return a
}

//MaxF max of two float64
func maxF(a, b float64) float64 {
	if a < b {
		return b
	}
	return a
}

//minF min of two float64
func minF(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
