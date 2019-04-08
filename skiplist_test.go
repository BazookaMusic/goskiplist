package goskiplist

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

const dataAmount = 1000                               // dataAmount of elements to insert
const nRoutinesAmount = 100                           // routines to start for concurrent tests
var nRoutinesToUse = min(dataAmount, nRoutinesAmount) // no reason to spawn more routines than inputs

var evenDataAmount = turnEven(dataAmount) // used for tests

const FAST = true // fast random generator

// helpers
func evalSort(arr []SkiplistItem) bool {
	if len(arr) == 0 {
		return true
	}
	prev := arr[0]
	for index := 1; index < len(arr); index++ {
		if arr[index].Less(prev) {
			fmt.Println("Items out of order:", arr[index], prev)
			return false
		}
		prev = arr[index]
	}
	return true
}

func debug(head *Skiplist) {
	for level := head.nLevels - 1; level >= 0; level-- {
		listHead := head.head

		for listHead != nil {
			fmt.Print(listHead.value, " ")
			listHead = listHead.next[level]
		}

		fmt.Println("nil")
	}

}

/* parallel inserters and removers */

func (head *Skiplist) Inserter(v int, wg *sync.WaitGroup) bool {
	defer wg.Done()
	for index := v * (dataAmount / nRoutinesToUse); index < (v+1)*(dataAmount/nRoutinesToUse); index++ {
		if !head.Insert(Int(index)) {
			return false
		}
	}
	return true
}

func (head *Skiplist) Remover(v int, wg *sync.WaitGroup) bool {

	defer wg.Done()
	for index := v * (dataAmount / nRoutinesToUse); index < (v+1)*(dataAmount/nRoutinesToUse); index++ {
		if !head.Remove(Int(index)) {
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

	var head = New(0.5, 30, FAST)

	//var wg sync.WaitGroup

	fmt.Println("Inserting numbers from 0 to", dataAmount-1)
	for index := 0; index < dataAmount; index++ {
		if !head.Insert(Int(index)) {
			t.Errorf("Could not insert item %d", index)
		}
	}

	for index := 0; index < dataAmount; index++ {
		if !head.Contains((Int(index))) {
			t.Errorf("Inserted number %d but not contained in Skiplist", index)
		}
	}

	sorted := head.ToSortedArray()
	ok := evalSort(sorted)

	if !ok {
		t.Errorf("Items out of order")
	}

	if head.nElements != dataAmount {
		t.Errorf("Skiplist should contain %d items but contains %d", dataAmount, head.nElements)
	}

	fmt.Println("OK!")
	fmt.Println("----------------------------------------")

}

func TestRemove(t *testing.T) {

	fmt.Println("---------------------------------------")
	fmt.Println("Sequential integer add, check and remove")
	fmt.Println("----------------------------------------")

	rand.Seed(time.Now().UTC().UnixNano())

	var head = New(0.5, 30, FAST)

	//var wg sync.WaitGroup

	fmt.Println("Inserting numbers from 0 to", dataAmount-1)
	for index := 0; index < dataAmount; index++ {
		if !head.Insert(Int(index)) {

		}
	}

	for index := 0; index < dataAmount; index++ {
		if !head.Contains((Int(index))) {
			t.Errorf("Inserted number %d but not contained in Skiplist", index)
		}
	}

	if head.nElements != dataAmount {
		t.Errorf("Skiplist should contain %d items but contains %d", dataAmount, head.nElements)
	}

	fmt.Println("Removing numbers from 0 to", dataAmount-1)
	amountRemoved := 0
	for index := 0; index < dataAmount; index++ {
		if !head.Remove((Int(index))) {
			t.Errorf("Inserted number %d but not contained in Skiplist", index)
		} else {
			amountRemoved++
		}

		if !(dataAmount-amountRemoved == head.nElements) {
			t.Errorf("Item %d reported removed but item count not updated", index)
		}
	}

	if !(head.nElements == 0) {
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

	var head = New(0.5, 30, FAST)

	//var wg sync.WaitGroup

	added := 0
	fmt.Println("Inserting", dataAmount, "random numbers")
	for index := 0; index < dataAmount; index++ {
		if head.Insert(Int(rand.Intn(dataAmount))) {
			added++
		}
	}

	amountRemoved := 0
	for index := 0; index < dataAmount; index++ {
		if head.Contains((Int(index))) {
			if !head.Remove((Int(index))) {
				t.Errorf("Number %d exists but could not be removed from Skiplist ", index)
			}

			amountRemoved++
		}
		// check order for some removes
		// if dataAmount > 100 && index%(dataAmount/100) == 0 {
		// 	sorted := head.ToSortedArray()
		// 	ok := evalSort(sorted)

		// 	if !ok {
		// 		t.Errorf("Items out of order after remove")
		// 	}
		// }

		if !(added-amountRemoved == head.nElements) {
			t.Errorf("Item %d reported removed but item count not updated", index)
		}
	}

	if !(head.nElements == 0) {
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

	var head = New(0.5, 30, FAST)

	var wg sync.WaitGroup

	wg.Add(nRoutinesToUse)
	fmt.Println("Spawing", nRoutinesToUse, "coroutines to insert ", dataAmount, "elements")
	for index := 0; index < nRoutinesToUse; index++ {
		if !head.Inserter(index, &wg) {
			t.Errorf("Could not insert item %d", index)
		}
	}

	wg.Wait()

	// lockless contains doesn't matter if run on one
	// or many coroutines
	for index := 0; index < dataAmount; index++ {
		if !head.Contains((Int(index))) {
			t.Errorf("Inserted number %d but not contained in Skiplist", index)
		}
	}

	sorted := head.ToSortedArray()
	ok := evalSort(sorted)

	if !ok {
		t.Errorf("Items out of order")
	}

	if head.nElements != dataAmount {
		t.Errorf("Skiplist should contain %d items but contains %d", dataAmount, head.nElements)
	}

	fmt.Println("OK!")
	fmt.Println("----------------------------------------")

}

func TestConcurrentInsertRemove(t *testing.T) {

	fmt.Println("--------------------------------------------")
	fmt.Println("Concurrent Sequential integer add and remove")
	fmt.Println("--------------------------------------------")

	rand.Seed(time.Now().UTC().UnixNano())

	var head = New(0.5, 30, FAST)

	var wg sync.WaitGroup

	wg.Add(nRoutinesToUse)
	fmt.Println("Spawing", nRoutinesToUse, "coroutines to insert ", dataAmount, "elements")
	for index := 0; index < nRoutinesToUse; index++ {
		if !head.Inserter(index, &wg) {
			t.Errorf("Could not insert item %d", index)
		}
	}

	wg.Wait()

	// lockless contains doesn't matter if run on one
	// or many coroutines
	for index := 0; index < dataAmount; index++ {
		if !head.Contains((Int(index))) {
			t.Errorf("Inserted number %d but not contained in Skiplist", index)
		}
	}

	wg.Add(nRoutinesToUse)
	fmt.Println("Spawing", nRoutinesToUse, "coroutines to remove ", dataAmount, "elements")
	for index := 0; index < nRoutinesToUse; index++ {
		if !head.Remover(index, &wg) {
			t.Errorf("Could not insert item %d", index)
		}
	}

	wg.Wait()

	if head.nElements != 0 {
		t.Errorf("Skiplist should be empty but contains %d elements", head.nElements)
	}

	fmt.Println("OK!")
	fmt.Println("----------------------------------------")

}

func TestConcurrentMixed(t *testing.T) {

	fmt.Println("---------------------------------------")
	fmt.Println("Mixed add and remove")
	fmt.Println("----------------------------------------")

	rand.Seed(time.Now().UTC().UnixNano())

	var head = New(0.5, 30, FAST)

	var wg sync.WaitGroup

	wg.Add(nRoutinesToUse)
	fmt.Println("Spawing", nRoutinesToUse, "coroutines to insert ", dataAmount, "elements from 0 to", dataAmount)
	fmt.Println("and spawing", nRoutinesToUse/2, "coroutines to remove elements from 0 to", dataAmount/2)
	for index := 0; index < nRoutinesToUse; index++ {
		if !head.Inserter(index, &wg) {
			t.Errorf("Could not insert item %d", index)
		}
	}

	wg.Wait()

	wg.Add(nRoutinesToUse / 2)
	fmt.Println("Spawing", nRoutinesToUse, "coroutines to remove ", dataAmount, "elements")
	for index := 0; index < nRoutinesToUse/2; index++ {
		if !head.Remover(index, &wg) {
			t.Errorf("Could not insert item %d", index)
		}
	}

	wg.Wait()

	// lockless contains doesn't matter if run on one
	// or many coroutines
	for index := 0; index < dataAmount/2; index++ {
		if head.Contains((Int(index))) {
			t.Errorf("%d should not be contained in Skiplist", index)
		}
	}

	if head.nElements != evenDataAmount/2 {
		t.Errorf("Skiplist should have %d elements but has %d", dataAmount/2, head.nElements)
	}

	fmt.Println("OK!")
	fmt.Println("----------------------------------------")

}

func TestConcurrentMixedModifyParams(t *testing.T) {

	fmt.Println("---------------------------------------")
	fmt.Println("Mixed add and remove, parameters modified concurrently")
	fmt.Println("----------------------------------------")

	rand.Seed(time.Now().UTC().UnixNano())

	var head = New(0.5, 30, FAST)

	var wg sync.WaitGroup

	wg.Add(nRoutinesToUse)
	fmt.Println("Spawing", nRoutinesToUse, "coroutines to insert ", dataAmount, "elements from 0 to", dataAmount)
	fmt.Println("and spawing", nRoutinesToUse/2, "coroutines to remove elements from 0 to", dataAmount/2)
	for index := 0; index < nRoutinesToUse; index++ {
		if !head.Inserter(index, &wg) {
			t.Errorf("Could not insert item %d", index)
		}
	}

	/* ACTUAL TEST */

	// change parameters concurrently
	go head.setProb(0.1)
	go head.setFastRandom(false)
	go head.setMaxLevels(12)

	/* ACTUAL TEST */

	wg.Wait()

	wg.Add(nRoutinesToUse / 2)
	fmt.Println("Spawing", nRoutinesToUse, "coroutines to remove ", dataAmount, "elements")
	for index := 0; index < nRoutinesToUse/2; index++ {
		if !head.Remover(index, &wg) {
			t.Errorf("Could not insert item %d", index)
		}
	}

	wg.Wait()

	// lockless contains doesn't matter if run on one
	// or many coroutines
	for index := 0; index < dataAmount/2; index++ {
		if head.Contains((Int(index))) {
			t.Errorf("%d should not be contained in Skiplist", index)
		}
	}

	if head.nElements != evenDataAmount/2 {
		t.Errorf("Skiplist should have %d elements but has %d", dataAmount/2, head.nElements)
	}

	fmt.Println("OK!")
	fmt.Println("----------------------------------------")

}

func TestUnion(t *testing.T) {
	fmt.Println("-------------------")
	fmt.Println("Skiplist union test")
	fmt.Println("-------------------")

	rand.Seed(time.Now().UTC().UnixNano())

	var head = New(0.5, 30, FAST)

	var mergedNew = New(0.5, 30, FAST)

	fmt.Println("Making first Skiplist")
	for index := 0; index < dataAmount; index++ {
		if !head.Insert(Int(index)) {
			t.Errorf("Could not insert item %d", index)
		}
	}

	for index := 0; index < dataAmount; index++ {
		if !head.Contains((Int(index))) {
			t.Errorf("Inserted number %d but not contained in Skiplist", index)
		}
	}

	var head2 = New(0.5, 30, FAST)

	fmt.Println("Making second Skiplist")
	for index := 0; index < 2*dataAmount; index += 2 {
		if !head2.Insert(Int(index)) {
			t.Errorf("Could not insert item %d", index)
		}
	}
	for index := 0; index < 2*dataAmount; index += 2 {
		if !head2.Contains((Int(index))) {
			t.Errorf("Inserted number %d but not contained in Skiplist", index)
		}
	}

	if head.nElements != dataAmount {
		t.Errorf("Skiplist should contain %d items but contains %d", dataAmount, head.nElements)
	}

	if head2.nElements != dataAmount {
		t.Errorf("Skiplist 2 should contain %d items but contains %d", dataAmount, head.nElements)
	}

	fmt.Println("Merging to new Skiplist...")

	var merged = mergedNew.Union(head, head2)

	if merged.nElements != dataAmount+dataAmount/2 {
		t.Errorf("Merged Skiplist should contain %d items but contains %d", dataAmount+dataAmount/2, merged.nElements)
	}

	head1Slice := head.ToSortedArray()
	head2Slice := head2.ToSortedArray()

	for _, item := range head1Slice {
		//fmt.Println(item)
		if !merged.Contains(item) {
			t.Errorf("First Skiplist contains %d but not contained in merged Skiplist", item.(Int))
		}
	}

	// for _, item := range head2Slice {
	// 	if !merged.Contains(item) {
	// 		t.Errorf("Second Skiplist contains %d but not contained in Skiplist", item.(int))
	// 	}
	// }

	merged = UnionSimple(head, head2)

	if merged.nElements != dataAmount+dataAmount/2 {
		t.Errorf("Merged Skiplist should contain %d items but contains %d", dataAmount+dataAmount/2, merged.nElements)
	}

	for _, item := range head1Slice {
		if !merged.Contains(item) {
			t.Errorf("First Skiplist contains %d but not contained in merged Skiplist", item.(Int))
		}
	}

	for _, item := range head2Slice {
		if !merged.Contains(item) {
			t.Errorf("Second Skiplist contains %d but not contained in Skiplist", item.(Int))
		}
	}

	fmt.Println("OK!")
	fmt.Println("----------------------------------------")
}

func TestIntersect(t *testing.T) {
	fmt.Println("-------------------")
	fmt.Println("Skiplist intersection test")
	fmt.Println("-------------------")

	rand.Seed(time.Now().UTC().UnixNano())

	var head = New(0.5, 30, FAST)

	var newSkiplist = New(0.5, 30, FAST)

	fmt.Printf("Making first Skiplist with elements %d to %d\n", 0, dataAmount)
	for index := 0; index < dataAmount; index++ {
		if !head.Insert(Int(index)) {
			t.Errorf("Could not insert item %d", index)
		}
	}

	for index := 0; index < dataAmount; index++ {
		if !head.Contains((Int(index))) {
			t.Errorf("Inserted number %d but not contained in Skiplist", index)
		}
	}

	var head2 = New(0.5, 30, FAST)

	fmt.Printf("Making second Skiplist with even elements %d to %d\n", 0, 2*dataAmount)
	for index := 0; index < 2*dataAmount; index += 2 {
		if !head2.Insert(Int(index)) {
			t.Errorf("Could not insert item %d", index)
		}
	}
	for index := 0; index < 2*dataAmount; index += 2 {
		if !head2.Contains((Int(index))) {
			t.Errorf("Inserted number %d but not contained in Skiplist", index)
		}
	}

	if head.nElements != dataAmount {
		t.Errorf("Skiplist should contain %d items but contains %d", dataAmount, head.nElements)
	}

	if head2.nElements != dataAmount {
		t.Errorf("Skiplist 2 should contain %d items but contains %d", dataAmount, head.nElements)
	}

	fmt.Println("Intersecting with new probabilities...")

	var intersected = newSkiplist.Intersection(head, head2)

	for index := 0; index < dataAmount; index += 2 {
		if !intersected.Contains(Int(index)) {
			t.Errorf("Number %d should be contained in Skiplist", index)
		}

	}

	fmt.Println("Intersecting and keeping Skiplist structures...")
	intersected = IntersectionSimple(head, head2)

	for index := 0; index < dataAmount; index += 2 {
		if !intersected.Contains(Int(index)) {
			t.Errorf("Number %d should be contained in Skiplist", index)
		}

	}

	fmt.Println("OK!")
	fmt.Println("----------------------------------------")
}

func BenchmarkInsert(b *testing.B) {
	rand.Seed(time.Now().UTC().UnixNano())

	var head = New(0.5, 30, FAST)

	b.ResetTimer()

	//var wg sync.WaitGroup

	for index := 0; index < b.N; index++ {
		if !head.Insert(Int(index)) {
			b.Errorf("Could not insert item %d", index)
		}
	}

}

func BenchmarkDelete(b *testing.B) {
	rand.Seed(time.Now().UTC().UnixNano())

	var head = New(0.5, 30, FAST)

	for index := 0; index < b.N; index++ {
		if !head.Insert(Int(index)) {
			b.Errorf("Could not insert item %d", index)
		}
	}

	b.ResetTimer()

	for index := 0; index < b.N; index++ {
		if !head.Remove(Int(index)) {
			b.Errorf("Could not insert item %d", index)
		}
	}

}

func BenchmarkSearch(b *testing.B) {
	rand.Seed(time.Now().UTC().UnixNano())

	var head = New(0.5, 30, FAST)

	for index := 0; index < b.N; index++ {
		if !head.Insert(Int(index)) {
			b.Errorf("Could not insert item %d", index)
		}
	}

	b.ResetTimer()

	for index := 0; index < b.N; index++ {
		if !head.Contains(Int(index)) {
			b.Errorf("Could not insert item %d", index)
		}
	}

}

func BenchmarkUnion(b *testing.B) {
	rand.Seed(time.Now().UTC().UnixNano())

	var head = New(0.5, 30, FAST)

	var head1 = New(0.5, 30, FAST)

	var union = New(0.5, 30, FAST)

	for index := 0; index < b.N; index++ {
		if !head.Insert(Int(index)) {
			b.Errorf("Could not insert item %d", index)
		}
	}

	b.ResetTimer()

	for index := 0; index < b.N; index++ {
		union = union.Union(head, head1)
	}

}

func BenchmarkUnionSimple(b *testing.B) {
	rand.Seed(time.Now().UTC().UnixNano())

	var head = New(0.5, 30, FAST)

	var head1 = New(0.5, 30, FAST)

	for index := 0; index < b.N; index++ {
		if !head.Insert(Int(index)) {
			b.Errorf("Could not insert item %d", index)
		}
	}

	for index := 0; index < b.N; index++ {
		if !head1.Insert(Int(index)) {
			b.Errorf("Could not insert item %d", index)
		}
	}

	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		UnionSimple(head, head1)
	}

}

func BenchmarkIntersection(b *testing.B) {
	rand.Seed(time.Now().UTC().UnixNano())

	var head = New(0.5, 30, FAST)

	var head1 = New(0.5, 30, FAST)

	var intersection = New(0.5, 30, FAST)

	for index := 0; index < b.N; index++ {
		if !head.Insert(Int(index)) {
			b.Errorf("Could not insert item %d", index)
		}
	}

	for index := b.N / 2; index < b.N; index++ {
		if !head1.Insert(Int(index)) {
			b.Errorf("Could not insert item %d", index)
		}
	}

	b.ResetTimer()

	for index := 0; index < b.N; index++ {
		intersection = intersection.Intersection(head, head1)
	}

}

func BenchmarkIntersectionSimple(b *testing.B) {
	rand.Seed(time.Now().UTC().UnixNano())

	var head = New(0.5, 30, FAST)

	var head1 = New(0.5, 30, FAST)

	for index := 0; index < b.N; index++ {
		if !head.Insert(Int(index)) {
			b.Errorf("Could not insert item %d", index)
		}
	}

	for index := b.N / 2; index < b.N; index++ {
		if !head1.Insert(Int(index)) {
			b.Errorf("Could not insert item %d", index)
		}
	}

	b.ResetTimer()

	for index := 0; index < b.N; index++ {
		IntersectionSimple(head, head1)
	}

}
