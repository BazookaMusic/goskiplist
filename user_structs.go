package main

// example struct setup for insertion to
// skiplist

type mystr struct {
	a int
	b bool
}

// Required

// do not delete int case,
// used for testing
func Less(a, b interface{}) bool {
	switch a.(type) {
	case mystr:
		a := a.(mystr)
		b := b.(mystr)
		return a.a < b.a
	case int:
		return a.(int) < b.(int)
	default:
		return false
	}
}

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
