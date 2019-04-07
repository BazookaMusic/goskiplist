[![Build Status](https://travis-ci.org/gerrish/goskiplist.svg?branch=master)](https://travis-ci.org/gerrish/goskiplist)

This is a thread safe implementation of a generic skiplist in Go. It is a direct implementation of the simple 
optimistic lazy Skiplist algorithm from the paper by Maurice Herlihy,Yossi Lev,Victor Luchangco and Nir Shavit.

[Link to paper](http://people.csail.mit.edu/shanir/publications/LazySkipList.pdf "the paper")

The implementation includes tests and the repository is connected with Travis for continuous integration.

The skiplist acts as a set and supports insert,contains,remove operations in O(logn) expected time and union and intersection operations in O(n + m) expected time. 

Example usage:
```golang
// initialise random number generator
rand.Seed(time.Now().UTC().UnixNano())

// allocate skiplist struct
head := new(Skiplist)


/* Initialise skiplist parameters */

// the first parameter is the probability of
// the bernoulli trials to choose the maximum list
// level

// the second parameter is the max amount of levels
// which will contain items. 
// SkiplistMaxLevel is the max amount allowed by the
// implementation (compile constant)

// the third parameter, if set to true,
// enables an optimised algorithm to
// generate the random levels but
// with a set probability of 0.5


head.InitSkiplist(0.5, SkiplistMaxLevel, false)

/* thread-safe insert,remove, contains */
dataAmount := 100

// insert

for index := 0; index < dataAmount; index++ {
    if !head.Insert(interface{}(index)) {
        // insertion failed, item already inserted
    }
}

// contains
for index := 0; index < dataAmount; index++ {
    if !head.Contains(interface{}(index)) {
        // item not in skiplist
    }
}

//remove
    for index := 0; index < dataAmount; index++ {
    if !head.Remove(interface{}(index)) {
        //  item not in skiplist
    }
}

/* will not corrupt the structure */
for index := 0; index < dataAmount; index++ {
    go head.Insert(interface{}(index))
    go head.Contains(interface{}(index))
    go head.Remove(interface{}(index))
    
}

/* Union */
var other = new(Skiplist)
other.InitSkiplist(0.5, 30, FAST)

for index := 0; index < 2*dataAmount; index++ {
    if !other.Insert(interface{}(index)) {
        // insertion failed, item already inserted
    }
}

/* new skiplist struct, set parameters */
var union = new(Skiplist)
union.InitSkiplist(0.5, 30, FAST)

/*the union items are inserted anew according to the parameters
of the initialized list
, complexity O(n + m)  */
union := union.Union(head, other)


/*the skiplists are directly merged level by level,
elements keep their former levels, no random calls.
faster but can lead to unbalanced skiplists.
New skiplist parameters set to defaults, see function
documentation. */
union := UnionSimple(head, other)

var intersection = new(Skiplist)
intersection.InitSkiplist(0.5, 30, FAST)

/*the intersection items are inserted anew according to the parameters
of the initialized list
, complexity O(n + m)  */
intersection := intersection.Intersection(head, other)


/*the skiplists are directly intersected level by level,
empty levels are omitted.
Elements keep their former levels, no random calls.
faster but can lead to unbalanced skiplists.
New skiplist parameters set to defaults, see function
documentation. */
intersection := IntersectionSimple(head, other)



```

