package main

import (
	"fmt"
	"math/rand"
	"time"
)

type skiplist_node struct {
	value int
	next  [30]*skiplist_node
	prev  [30]*skiplist_node
}
type skiplist struct {
	n_levels   int
	head       *skiplist_node
	n_elements int
	prob       float64
	max_levels int
}

func (head *skiplist) Height() int {
	return head.n_levels
}

func (head *skiplist) Len() int {
	return head.n_elements
}

type NotFoundErr string

func (err NotFoundErr) Error() string {
	return "Element not found"
}
func Init_skiplist(prob float64, max_levels int) skiplist {
	var initial skiplist
	initial.n_levels = 1
	initial.prob = prob
	initial.max_levels = max_levels

	var head skiplist_node
	head.value = -1

	initial.n_elements = 0

	initial.head = &head

	return initial
}

func debug(a skiplist) {
	level := a.n_levels - 1
	finger := a.head

	for ; level >= 0; level-- {
		counter := 0
		if finger.next[level] == nil {
			fmt.Print(nil)
		}
		for node := finger.next[level]; node != nil; node = node.next[level] {
			counter++
			fmt.Print(" ", node.value)
		}
		fmt.Println(" nil", counter)
	}
}

func (list *skiplist) ToSortedArray() []int {
	arr := make([]int, list.n_elements, list.n_elements)
	counter := 0
	for current_node := list.head.next[0]; current_node != nil; current_node = current_node.next[0] {
		arr[counter] = current_node.value
		//fmt.Print(arr[counter], current_node.value, " ")
		counter++
	}

	//fmt.Println(arr)

	return arr

}

func (head *skiplist) Find(val int) (*skiplist_node, int) {
	curr := head.head
	level := head.n_levels - 1
	// vertically
	for ; level >= 0; level-- {
		// horizontally
		for ; curr.next[level] != nil && curr.next[level].value < val; curr = curr.next[level] {
		}
		//found something or have to go down

		// is the next element what I seek
		if curr.next[level] != nil && curr.next[level].value == val {
			return curr.next[level], level
		}
	}
	// not found
	return nil, 0
}

func (head *skiplist) Remove(val int) error {

	node, level := head.Find(val)

	if node == nil {
		// never found
		return new(NotFoundErr) // error
	} else {
		// traverse down and delete nodes
		for ; level >= 0; level-- {

			node.prev[level].next[level] = node.next[level]
			if node.next[level] != nil {
				node.next[level].prev[level] = node.prev[level]
			}

		}

		// removed properly
		head.n_elements--
		return nil

	}

}

func (head *skiplist) Insert(val int) int {

	head.n_elements++

	levels := coin_tosses(head.prob, head.max_levels)
	//fmt.Println("element ", val, levels)

	// expand height
	if levels > head.n_levels {
		head.n_levels = levels // new # of levels
	}

	// new node created
	new_val := new(skiplist_node)
	new_val.value = val

	/*
		//fmt.Println("insertion point", ins_point)
		prev = head.head
		// traverse vertically
		for level := head.n_levels - 1; level >= 0; level-- {


			// traverse horizontally to find insertion point
			for current_node = ins_point.next[level]; current_node != nil && current_node.value < val; prev, current_node = current_node, current_node.next[level] {
			}
		}
	*/

	curr := head.head
	level := head.n_levels - 1
	// vertically
	for ; level >= 0; level-- {
		// horizontally
		//fmt.Println("a")
		for ; curr.next[level] != nil && curr.next[level].value < val; curr = curr.next[level] {
			//fmt.Println(curr)
			//fmt.Println(curr)
		}
		//found something or have to go down
		//fmt.Println("proceed")
		//fmt.Println(curr)
		// insert if level is less or equal to random tosses
		if level <= levels {
			new_val.next[level] = curr.next[level]
			new_val.prev[level] = curr
			curr.next[level] = new_val

		}

		//fmt.Println(level)

	}

	return 0

}

func coin_tosses(prob float64, max_levels int) int {
	var counter int = 1
	res := rand.Float64()
	for res > prob && counter < max_levels {
		counter++
		res = rand.Float64()
	}

	return counter

}

func eval_sort(arr []int) {
	prev := arr[0]
	for index := 1; index < len(arr); index++ {
		if arr[index] < prev {
			fmt.Println(prev, arr[index], index)
		}
		prev = arr[index]
	}

	fmt.Println(len(arr))
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	var head skiplist
	head = Init_skiplist(0.5, 20)

	var arr []int = make([]int, 10000000)
	for index := 0; index < 1000000; index++ {
		//head.Insert(rand.Intn(1234500))
		head.Insert(index)
		arr[index] = index
	}

	//sort.Ints(arr)
	head.Remove(1)
	head.Remove(2)

	sorted := head.ToSortedArray()
	eval_sort(sorted)
	fmt.Println(head.Len())

	fmt.Println(sorted[0], sorted[1])

	//debug(head)
}
