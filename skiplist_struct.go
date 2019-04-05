package main

import "sync"

// SkiplistMaxLevel maximum levels allocated for each Skiplist
// next pointer arrays are of constant size
const SkiplistMaxLevel = 30

//MinProb Minimum probability of bernoulli trial success
//for random function
const MinProb = 0.01

type skiplistNode struct {
	value       interface{}
	next        [SkiplistMaxLevel]*skiplistNode
	marked      bool
	fullyLinked bool
	mux         sync.Mutex
	topLevel    int
}

/*Skiplist : The Skiplist structure, must be initialised before use. */
type Skiplist struct {
	nLevels    int
	head       *skiplistNode
	nElements  int
	prob       float64
	maxLevels  int
	lock       sync.Mutex
	fastRandom bool
}
