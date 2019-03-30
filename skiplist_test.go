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
		if index%(AMOUNT/100) == 0 {
			sorted := head.ToSortedArray()
			ok := eval_sort(sorted)

			if !ok {
				t.Errorf("Items out of order after remove")
			}
		}

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
