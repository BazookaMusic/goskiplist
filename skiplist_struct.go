package main

import "sync"

const SKIPLIST_MAX_LEVEL = 30

type skiplist_node struct {
	value        interface{}
	next         [SKIPLIST_MAX_LEVEL]*skiplist_node
	marked       bool
	fully_linked bool
	mux          sync.Mutex
	top_level    int
}

type skiplist struct {
	n_levels    int
	head        *skiplist_node
	n_elements  int
	prob        float64
	max_levels  int
	lock        sync.Mutex
	fast_random bool
}

type NotFoundErr string

func (err NotFoundErr) Error() string {
	return "Element not found"
}
