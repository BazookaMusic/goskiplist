package main

// example struct setup for insertion to
// Skiplist

type mystr struct {
	a int
	b bool
}

// Less : Node comparison function, should be
// set for user struct. Returns true if a is less than b
func (a mystr) Less(b SkiplistItem) bool {
	return false
}

// Equals : Node comparison function, should be
// set for user struct. Returns true if a equals b
func (a mystr) Equals(b SkiplistItem) bool {
	return false
}

// Required for tests

//Int : an integer
type Int int

// Less : Node comparison function for Int, should be
// set for user struct. Returns true if a is less than b
func (a Int) Less(b SkiplistItem) bool {
	c, ok := b.(Int)

	return ok && a < c
}

// Equals : Node comparison function for Int, should be
// set for user struct. Returns true if a equals b
func (a Int) Equals(b SkiplistItem) bool {
	c, ok := b.(Int)

	return ok && a == c
}
