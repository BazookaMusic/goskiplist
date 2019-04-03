package main

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

const AMOUNT = 10000   // amount of elements to insert
const N_ROUTINES = 100 // routines to start for concurrent tests
const FAST = true      // fast random generator

// helpers
func eval_sort(arr []interface{}) bool {
	if len(arr) == 0 {
		return true
	}
	prev := arr[0]
	for index := 1; index < len(arr); index++ {
		if Less(arr[index], prev) {
			fmt.Println("Items out of order:", arr[index], prev)
			return false
		}
		prev = arr[index]
	}
	return true
}

func debug(head *skiplist) {
	for level := head.n_levels - 1; level >= 0; level-- {
		list_head := head.head

		for list_head != nil {
			fmt.Print(list_head.value, " ")
			list_head = list_head.next[level]
		}

		fmt.Println("nil")
	}

}

func (head *skiplist) Inserter(v int, wg *sync.WaitGroup) bool {
	defer wg.Done()
	for index := v * (AMOUNT / N_ROUTINES); index < (v+1)*(AMOUNT/N_ROUTINES); index++ {
		if !head.Insert(interface{}(index)) {
			return false
		}
	}
	return true
}

func (head *skiplist) Remover(v int, wg *sync.WaitGroup) bool {

	defer wg.Done()
	for index := v * (AMOUNT / N_ROUTINES); index < (v+1)*(AMOUNT/N_ROUTINES); index++ {
		if !head.Remove(interface{}(index)) {
			return false
		}
	}
	return true

}

// basic functionality tests
func TestInsert(t *testing.T) {

	fmt.Println("---------------------------------------")
	fmt.Println("Sequential integer add and test order")
	fmt.Println("----------------------------------------")

	rand.Seed(time.Now().UTC().UnixNano())

	var head *skiplist = new(skiplist)
	head.Init_skiplist(0.5, 30, FAST)

	//var wg sync.WaitGroup

	fmt.Println("Inserting numbers from 0 to", AMOUNT-1)
	for index := 0; index < AMOUNT; index++ {
		if !head.Insert(interface{}(index)) {
			t.Errorf("Could not insert item %d", index)
		}
	}

	for index := 0; index < AMOUNT; index++ {
		if !head.Contains((interface{}(index))) {
			t.Errorf("Inserted number %d but not contained in skiplist", index)
		}
	}

	sorted := head.ToSortedArray()
	ok := eval_sort(sorted)

	if !ok {
		t.Errorf("Items out of order")
	}

	if head.n_elements != AMOUNT {
		t.Errorf("Skiplist should contain %d items but contains %d", AMOUNT, head.n_elements)
	}

	fmt.Println("OK!")
	fmt.Println("----------------------------------------")

}

func TestRemove(t *testing.T) {

	fmt.Println("---------------------------------------")
	fmt.Println("Sequential integer add, check and remove")
	fmt.Println("----------------------------------------")

	rand.Seed(time.Now().UTC().UnixNano())

	var head *skiplist = new(skiplist)
	head.Init_skiplist(0.5, 20, FAST)

	//var wg sync.WaitGroup

	fmt.Println("Inserting numbers from 0 to", AMOUNT-1)
	for index := 0; index < AMOUNT; index++ {
		if !head.Insert(interface{}(index)) {

		}
	}

	for index := 0; index < AMOUNT; index++ {
		if !head.Contains((interface{}(index))) {
			t.Errorf("Inserted number %d but not contained in skiplist", index)
		}
	}

	if head.n_elements != AMOUNT {
		t.Errorf("Skiplist should contain %d items but contains %d", AMOUNT, head.n_elements)
	}

	fmt.Println("Removing numbers from 0 to", AMOUNT-1)
	amount_removed := 0
	for index := 0; index < AMOUNT; index++ {
		if !head.Remove((interface{}(index))) {
			t.Errorf("Inserted number %d but not contained in skiplist", index)
		} else {
			amount_removed++
		}

		if !(AMOUNT-amount_removed == head.n_elements) {
			t.Errorf("Item %d reported removed but item count not updated", index)
		}
	}

	if !(head.n_elements == 0) {
		t.Errorf("Skiplist should be empty")
	}

	fmt.Println("OK!")
	fmt.Println("----------------------------------------")

}

func TestRandOperation(t *testing.T) {
	fmt.Println("------------------------------------")
	fmt.Println("Random integer add, check and remove")
	fmt.Println("------------------------------------")
	rand.Seed(time.Now().UTC().UnixNano())

	var head *skiplist = new(skiplist)
	head.Init_skiplist(0.5, 20, FAST)

	//var wg sync.WaitGroup

	added := 0
	fmt.Println("Inserting", AMOUNT, "random numbers")
	for index := 0; index < AMOUNT; index++ {
		if head.Insert(interface{}(rand.Intn(AMOUNT))) {
			added++
		}
	}

	amount_removed := 0
	for index := 0; index < AMOUNT; index++ {
		if head.Contains((interface{}(index))) {
			if !head.Remove((interface{}(index))) {
				t.Errorf("Number %d exists but could not be removed from skiplist ", index)
			}

			amount_removed++
		}
		// check order for some removes
		// if AMOUNT > 100 && index%(AMOUNT/100) == 0 {
		// 	sorted := head.ToSortedArray()
		// 	ok := eval_sort(sorted)

		// 	if !ok {
		// 		t.Errorf("Items out of order after remove")
		// 	}
		// }

		if !(added-amount_removed == head.n_elements) {
			t.Errorf("Item %d reported removed but item count not updated", index)
		}
	}

	if !(head.n_elements == 0) {
		t.Errorf("Skiplist should be empty")
	}

	fmt.Println("OK!")
	fmt.Println("----------------------------------------")

}

func TestConcurrentInsertAndOrder(t *testing.T) {

	fmt.Println("--------------------------------------")
	fmt.Println("Sequential integer add and test order")
	fmt.Println("-------------------------------------")

	rand.Seed(time.Now().UTC().UnixNano())

	var head *skiplist = new(skiplist)
	head.Init_skiplist(0.5, 20, FAST)

	var wg sync.WaitGroup

	wg.Add(N_ROUTINES)
	fmt.Println("Spawing", N_ROUTINES, "coroutines to insert ", AMOUNT, "elements")
	for index := 0; index < N_ROUTINES; index++ {
		if !head.Inserter(index, &wg) {
			t.Errorf("Could not insert item %d", index)
		}
	}

	wg.Wait()

	// lockless contains doesn't matter if run on one
	// or many coroutines
	for index := 0; index < AMOUNT; index++ {
		if !head.Contains((interface{}(index))) {
			t.Errorf("Inserted number %d but not contained in skiplist", index)
		}
	}

	sorted := head.ToSortedArray()
	ok := eval_sort(sorted)

	if !ok {
		t.Errorf("Items out of order")
	}

	if head.n_elements != AMOUNT {
		t.Errorf("Skiplist should contain %d items but contains %d", AMOUNT, head.n_elements)
	}

	fmt.Println("OK!")
	fmt.Println("----------------------------------------")

}

func TestConcurrentInsertRemove(t *testing.T) {

	fmt.Println("--------------------------------------------")
	fmt.Println("Concurrent Sequential integer add and remove")
	fmt.Println("--------------------------------------------")

	rand.Seed(time.Now().UTC().UnixNano())

	var head *skiplist = new(skiplist)
	head.Init_skiplist(0.5, 20, FAST)

	var wg sync.WaitGroup

	wg.Add(N_ROUTINES)
	fmt.Println("Spawing", N_ROUTINES, "coroutines to insert ", AMOUNT, "elements")
	for index := 0; index < N_ROUTINES; index++ {
		if !head.Inserter(index, &wg) {
			t.Errorf("Could not insert item %d", index)
		}
	}

	wg.Wait()

	// lockless contains doesn't matter if run on one
	// or many coroutines
	for index := 0; index < AMOUNT; index++ {
		if !head.Contains((interface{}(index))) {
			t.Errorf("Inserted number %d but not contained in skiplist", index)
		}
	}

	wg.Add(N_ROUTINES)
	fmt.Println("Spawing", N_ROUTINES, "coroutines to remove ", AMOUNT, "elements")
	for index := 0; index < N_ROUTINES; index++ {
		if !head.Remover(index, &wg) {
			t.Errorf("Could not insert item %d", index)
		}
	}

	wg.Wait()

	if head.n_elements != 0 {
		t.Errorf("Skiplist should be empty but contains %d elements", head.n_elements)
	}

	fmt.Println("OK!")
	fmt.Println("----------------------------------------")

}

func TestConcurrentMixed(t *testing.T) {

	fmt.Println("---------------------------------------")
	fmt.Println("Mixed add and remove")
	fmt.Println("----------------------------------------")

	rand.Seed(time.Now().UTC().UnixNano())

	var head *skiplist = new(skiplist)
	head.Init_skiplist(0.5, 20, FAST)

	var wg sync.WaitGroup

	wg.Add(N_ROUTINES)
	fmt.Println("Spawing", N_ROUTINES, "coroutines to insert ", AMOUNT, "elements from 0 to", AMOUNT)
	fmt.Println("and spawing", N_ROUTINES/2, "coroutines to remove elements from 0 to", AMOUNT/2)
	for index := 0; index < N_ROUTINES; index++ {
		if !head.Inserter(index, &wg) {
			t.Errorf("Could not insert item %d", index)
		}
	}

	wg.Wait()

	wg.Add(N_ROUTINES / 2)
	fmt.Println("Spawing", N_ROUTINES, "coroutines to remove ", AMOUNT, "elements")
	for index := 0; index < N_ROUTINES/2; index++ {
		if !head.Remover(index, &wg) {
			t.Errorf("Could not insert item %d", index)
		}
	}

	wg.Wait()

	// lockless contains doesn't matter if run on one
	// or many coroutines
	for index := 0; index < AMOUNT/2; index++ {
		if head.Contains((interface{}(index))) {
			t.Errorf("%d should not be contained in skiplist", index)
		}
	}

	if head.n_elements != AMOUNT/2 {
		t.Errorf("Skiplist should have %d elements but has %d", AMOUNT/2, head.n_elements)
	}

	fmt.Println("OK!")
	fmt.Println("----------------------------------------")

}

func TestMerge(t *testing.T) {
	fmt.Println("-------------------")
	fmt.Println("Skiplist union test")
	fmt.Println("-------------------")

	rand.Seed(time.Now().UTC().UnixNano())

	var head *skiplist = new(skiplist)
	head.Init_skiplist(0.5, 30, FAST)

	//var wg sync.WaitGroup

	fmt.Println("Making first skiplist")
	for index := 0; index < AMOUNT; index++ {
		if !head.Insert(interface{}(index)) {
			t.Errorf("Could not insert item %d", index)
		}
	}

	for index := 0; index < AMOUNT; index++ {
		if !head.Contains((interface{}(index))) {
			t.Errorf("Inserted number %d but not contained in skiplist", index)
		}
	}

	var head2 *skiplist = new(skiplist)
	head2.Init_skiplist(0.5, 30, FAST)

	fmt.Println("Making second skiplist")
	for index := 0; index < 2*AMOUNT; index += 2 {
		if !head2.Insert(interface{}(index)) {
			t.Errorf("Could not insert item %d", index)
		}
	}
	for index := 0; index < 2*AMOUNT; index += 2 {
		if !head2.Contains((interface{}(index))) {
			t.Errorf("Inserted number %d but not contained in skiplist", index)
		}
	}

	if head.n_elements != AMOUNT {
		t.Errorf("Skiplist should contain %d items but contains %d", AMOUNT, head.n_elements)
	}

	if head2.n_elements != AMOUNT {
		t.Errorf("Skiplist 2 should contain %d items but contains %d", AMOUNT, head.n_elements)
	}

	fmt.Println("Merging...")

	var merged *skiplist = Union(head, head2, true)

	if merged.n_elements != AMOUNT+AMOUNT/2 {
		t.Errorf("Merged skiplist should contain %d items but contains %d", AMOUNT+AMOUNT/2, merged.n_elements)
	}

	head1_slice := head.ToSortedArray()
	head2_slice := head2.ToSortedArray()

	for _, item := range head1_slice {
		if !merged.Contains(item) {
			t.Errorf("First skiplist contains %d but not contained in merged skiplist", item.(int))
		}
	}

	for _, item := range head2_slice {
		if !merged.Contains(item) {
			t.Errorf("Second skiplist contains %d but not contained in skiplist", item.(int))
		}
	}

	fmt.Println("OK!")
	fmt.Println("----------------------------------------")
}

func TestIntersect(t *testing.T) {
	fmt.Println("-------------------")
	fmt.Println("Skiplist union test")
	fmt.Println("-------------------")

	rand.Seed(time.Now().UTC().UnixNano())

	var head *skiplist = new(skiplist)
	head.Init_skiplist(0.5, 30, FAST)

	//var wg sync.WaitGroup

	fmt.Println("Making first skiplist")
	for index := 0; index < AMOUNT; index++ {
		if !head.Insert(interface{}(index)) {
			t.Errorf("Could not insert item %d", index)
		}
	}

	for index := 0; index < AMOUNT; index++ {
		if !head.Contains((interface{}(index))) {
			t.Errorf("Inserted number %d but not contained in skiplist", index)
		}
	}

	var head2 *skiplist = new(skiplist)
	head2.Init_skiplist(0.5, 30, FAST)

	fmt.Println("Making second skiplist")
	for index := 0; index < 2*AMOUNT; index += 2 {
		if !head2.Insert(interface{}(index)) {
			t.Errorf("Could not insert item %d", index)
		}
	}
	for index := 0; index < 2*AMOUNT; index += 2 {
		if !head2.Contains((interface{}(index))) {
			t.Errorf("Inserted number %d but not contained in skiplist", index)
		}
	}

	if head.n_elements != AMOUNT {
		t.Errorf("Skiplist should contain %d items but contains %d", AMOUNT, head.n_elements)
	}

	if head2.n_elements != AMOUNT {
		t.Errorf("Skiplist 2 should contain %d items but contains %d", AMOUNT, head.n_elements)
	}

	fmt.Println("Intersecting...")

	var intersected *skiplist = Intersection(head, head2, true)
	fmt.Println("A")
	debug(head)
	fmt.Println("B")
	debug(head2)
	fmt.Println("INTER")
	debug(intersected)

	// if merged.n_elements != AMOUNT+AMOUNT/2 {
	// 	t.Errorf("Merged skiplist should contain %d items but contains %d", AMOUNT+AMOUNT/2, merged.n_elements)
	// }

	// head1_slice := head.ToSortedArray()
	// head2_slice := head2.ToSortedArray()

	// for _, item := range head1_slice {
	// 	if !merged.Contains(item) {
	// 		t.Errorf("First skiplist contains %d but not contained in merged skiplist", item.(int))
	// 	}
	// }

	// for _, item := range head2_slice {
	// 	if !merged.Contains(item) {
	// 		t.Errorf("Second skiplist contains %d but not contained in skiplist", item.(int))
	// 	}
	// }

	fmt.Println("OK!")
	fmt.Println("----------------------------------------")
}
