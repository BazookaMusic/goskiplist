package main

//Min min of two ints
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

//Max max of two ints
func Max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

// TurnEven if int is even return it,
// else return next even int
func TurnEven(a int) int {
	if a&1 == 1 {
		return a + 1
	}
	return a
}
