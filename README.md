[![Go Report Card](https://goreportcard.com/badge/github.com/gerrish/goskiplist)](https://goreportcard.com/report/github.com/gerrish/goskiplist)
[![Build Status](https://travis-ci.org/gerrish/goskiplist.svg?branch=master)](https://travis-ci.org/gerrish/goskiplist)

This is a thread safe implementation of a generic skiplist in Go. It is a direct implementation of the simple 
optimistic lazy Skiplist algorithm from the paper by Maurice Herlihy,Yossi Lev of Brown University & Sun Microsystems Laboratories,Victor Luchangco of Sun Microsystems Laboratories  and Nir Shavit of Tel-Aviv University & Sun Microsystems Laboratories.

[Link to paper](http://people.csail.mit.edu/shanir/publications/LazySkipList.pdf "the paper")

The implementation includes tests and the repository is connected with Travis for continuous integration.

The skiplist acts as a set and supports insert,contains,remove operations in O(logn) expected time and union and intersection operations in O(n + m) expected time. 

Example usage:
```golang
package main

import (
	"math/rand"
	"time"

	sl "github.com/gerrish/skiplist"
)

/* items must be converted to interface values,
   thus types must support Less,Equals methods */

// Int and its functions are already defined for
// convenience. See comparison_structs.go

//Int : an integer
type Int int

// type must support Less(a SkiplistItem) bool and Equals(a SkiplistItem) bool

// Less : Node comparison function for Int, should be
// set for user struct. Returns true if a is less than b
func (a Int) Less(b sl.SkiplistItem) bool {
	b, ok := b.(Int)

	return ok && int(a) < int(b.(Int))
}

// Equals : Node comparison function for Int, should be
// set for user struct. Returns true if a equals b
func (a Int) Equals(b sl.SkiplistItem) bool {
	b, ok := b.(Int)

	return ok && a == b
}

func main() {

	// initialise random number generator
	rand.Seed(time.Now().UTC().UnixNano())

	/* Initialise skiplist parameters */

	// the first parameter is the probability of
	// the bernoulli trials to choose the maximum list
	// level

	// the second parameter is the max amount of levels
	// which will contain items.
	// SkiplistMaxLevel is the max amount allowed by the
	// implementation (compile constant)

	// the third parameter, if set to FAST,
	// enables an optimised algorithm to
	// generate the random levels but
	// with a set probability of 0.5
	// If set to VARIABLE, allows variable probability
	// on random level generation, but is slower

	head := sl.New(0.5, 30, sl.FAST)

	/* thread-safe insert,remove, contains */
	dataAmount := 100

	// insert

	for index := 0; index < dataAmount; index++ {
		/* type conversion */
		if !head.Insert(Int(index)) {
			// insertion failed, item already inserted
		}
	}

	// contains
	for index := 0; index < dataAmount; index++ {
		if !head.Contains(Int(index)) {
			// item not in skiplist
		}
	}

	// get 
	item := head.Get(Int(dataAmount/3)) // get item with item.Equals(15) == true,
							 // useful to use set as a map
							 // nil if not contained
	

	//remove
	for index := 0; index < dataAmount; index++ {
		if !head.Remove(Int(index)) {
			//  item not in skiplist
		}
	}

	/* will not corrupt the structure */
	for index := 0; index < dataAmount; index++ {
		go head.Insert(Int(index))
		go head.Contains(Int(index))
		go head.Remove(Int(index))

	}

	/* Union */
	var other = sl.New(0.5, 30, sl.FAST)

	for index := 0; index < 2*dataAmount; index++ {
		if !other.Insert(Int(index)) {
			// insertion failed, item already inserted
		}
	}

	/* new skiplist struct, set parameters */
	var union = sl.New(0.5, 30, sl.FAST)

	/*the union items are inserted anew according to the parameters
	  of the initialized list
	  , complexity O(n + m)  */
	union = union.Union(head, other)

	/*the skiplists are directly merged level by level,
	  elements keep their former levels, no random calls.
	  sl.FASTer but can lead to unbalanced skiplists.
	  New skiplist parameters set to defaults, see function
	  documentation. */
	union = sl.UnionSimple(head, other)

	var intersection = sl.New(0.5, 30, sl.FAST)

	/*the intersection items are inserted anew according to the parameters
	  of the initialized list
	  , complexity O(n + m)  */
	intersection = intersection.Intersection(head, other)

	/*the skiplists are directly intersected level by level,
	  empty levels are omitted.
	  Elements keep their former levels, no random calls.
	  faster but can lead to unbalanced skiplists.
	  New skiplist parameters set to defaults, see function
	  documentation. */
	intersection = sl.IntersectionSimple(head, other)

}

```

Some performance results on an i5 3570K at 4.5Ghz and 16GB of ram at 1600mhz:
![Benchmark results](https://i.imgur.com/QyWd4ji.png)

