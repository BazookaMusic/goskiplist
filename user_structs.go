package main

// example struct setup for insertion to
// Skiplist

type mystr struct {
	a int
	b bool
}

// Required

// Less : Node comparison function, should be
// set for user struct. Returns true if a is less than b
func Less(a, b interface{}) bool {
	switch a.(type) {
	case mystr:
		a := a.(mystr)
		b := b.(mystr)
		return a.a < b.a
	// Do not delete int case, used for testing
	case int:
		return a.(int) < b.(int)
	case float64:
		return a.(float64) < b.(float64)
	default:
		return false
	}
}

// Equals : Node comparison function, should be
// set for user struct. Returns true if a equals b
func Equals(a, b interface{}) bool {
	switch a.(type) {
	case mystr:
		a := a.(mystr)
		b := b.(mystr)
		return a.a == b.a
	case int:
		return a.(int) == b.(int)
	default:
		return false

	}
}
