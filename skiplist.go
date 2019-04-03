package main

import (
	"fmt"
	"time"
)

var randtime time.Duration = 0

func (head *skiplist) Height() int {
	/* current max level */
	defer head.lock.Unlock()
	head.lock.Lock()
	return head.n_levels
}

func (head *skiplist) Len() int {
	/* current amount of inserted elements */
	defer head.lock.Unlock()
	head.lock.Lock()
	return head.n_elements
}

func (initial *skiplist) Init_skiplist(prob float64, max_levels int, fast_random bool) {
	/* Initialise skiplist */

	if prob < 0 {
		prob = 0.5
		fmt.Println("Init: Probability given less than zero, set to 0.5 instead")
	}

	if max_levels > SKIPLIST_MAX_LEVEL {
		fmt.Println("Init: Max level given more than supported amount of",
			SKIPLIST_MAX_LEVEL, " setting to ", SKIPLIST_MAX_LEVEL, "instead")
	}

	initial.n_levels = 1
	initial.prob = prob
	initial.max_levels = max_levels
	initial.fast_random = fast_random

	var head *skiplist_node = new(skiplist_node)
	head.fully_linked = true
	head.marked = false

	initial.n_elements = 0

	initial.head = head
}

func (list *skiplist) ToSortedArray() []interface{} {
	/* make a sorted array out of the skiplist
	   returns the lowest level               */
	arr := make([]interface{}, list.n_elements, list.n_elements)
	counter := 0
	for current_node := list.head.next[0]; current_node != nil; current_node = current_node.next[0] {
		arr[counter] = current_node.value
		counter++
	}

	return arr

}

func (head *skiplist) FindNextLowest(val interface{}) (node *skiplist_node) {
	/* Find where the element should be
	and return it's successor on the first level.
	Returns the element or
	-1 when not found */

	// could be modified by inserts
	head.lock.Lock()
	level := head.n_levels - 1
	head.lock.Unlock()
	// much faster than starting at max

	pred := head.head

	var curr *skiplist_node

	// traverse vertically
	for ; level >= 0; level-- {
		// horizontally
		curr = pred.next[level]
		for curr != nil && Less(curr.value, val) {
			pred = curr
			curr = pred.next[level]
		}

		// next of where it should be
		if curr != nil && Equals(curr.value, val) {
			break
		}

	}

	return curr
}

func (head *skiplist) Find(val interface{}, prev, next []*skiplist_node) (found_level int) {
	/* Find where the element should be
	and return the first level where it was found and
	next and previous elements for every level.
	Returns the first level where it was found or
	-1 when not found */

	// could be modified by inserts
	head.lock.Lock()
	level := head.n_levels - 1
	head.lock.Unlock()
	// much faster than starting at max

	pred := head.head
	found_level = -1
	var curr *skiplist_node

	// traverse vertically
	for ; level >= 0; level-- {
		// horizontally
		curr = pred.next[level]
		for curr != nil && Less(curr.value, val) {
			pred = curr
			curr = pred.next[level]
		}

		// next of where it should be
		if curr != nil && Equals(curr.value, val) && found_level == -1 {
			found_level = level
		}

		// previous of where the item should be
		prev[level] = pred
		next[level] = curr

	}

	return found_level
}

func (head *skiplist) Contains(val interface{}) bool {
	/* same function as find but returns
	as soon as item is found, ignoring below levels
	and does not return prev,next */

	head.lock.Lock()
	level := head.n_levels - 1
	head.lock.Unlock()

	pred := head.head
	var curr *skiplist_node
	// vertically
	for ; level >= 0; level-- {
		// horizontally
		curr = pred.next[level]
		for curr != nil && Less(curr.value, val) {
			pred = curr
			curr = pred.next[level]
		}
		//found something or have to go down

		// is the next element what I seek
		if curr != nil && Equals(curr.value, val) {
			node := curr
			return node.fully_linked && !node.marked
		}
	}
	// not found
	return false
}

func (head *skiplist) Insert(v interface{}) bool {
	// insert element

	// highest level of insertion
	// the head.fast property should not be modified after init
	top_level := coin_tosses(head.prob, head.max_levels, head.fast_random)

	// check if list must become taller
	head.lock.Lock()
	if top_level > head.n_levels {
		head.n_levels = top_level
	}
	head.lock.Unlock()

	// buffers to store prev and next pointers
	var prev, next []*skiplist_node
	prev = make([]*skiplist_node, SKIPLIST_MAX_LEVEL)
	next = make([]*skiplist_node, SKIPLIST_MAX_LEVEL)

	for {

		// find insertion point and previous and next nodes
		found_level := head.Find(v, prev, next)

		// already in skiplist
		if found_level != -1 {

			// should be the node with value v
			nodeFound := next[found_level]
			// if node is not set for removal
			if !nodeFound.marked {
				// wait until stable
				for !nodeFound.fully_linked {
				}
				//don't insert
				return false
			}
			// try again
			continue

		}
		// highest level locked
		highest_locked := -1
		var pred, succ *skiplist_node
		var prevPred *skiplist_node = nil

		valid := true

		// validate that new node can be added
		// by checking previous and next nodes
		for level := 0; valid && level < top_level; level++ {

			pred = prev[level]
			succ = next[level]

			// avoid locking same node twice
			// if two or more levels
			// connected to same node
			if pred != prevPred {
				pred.mux.Lock()

				highest_locked = level
				prevPred = pred
			}

			// can the insertion proceed
			// node is locked so we can check next
			valid = !pred.marked && (succ == nil || !succ.marked) && pred.next[level] == succ
		}

		// cannot add
		if !valid {
			// unlock to try again
			prevPred = nil
			for i := highest_locked; i >= 0; i-- {
				if prevPred != prev[i] {
					prev[i].mux.Unlock()
				}
				prevPred = prev[i]

			}
			// restart attempt
			continue
		}

		// try to add new node
		newNode := new(skiplist_node)
		newNode.value = v
		newNode.top_level = top_level - 1
		newNode.marked = false

		for level := 0; level < top_level; level++ {

			newNode.next[level] = next[level]
			prev[level].next[level] = newNode
		}
		// new node is ok
		newNode.fully_linked = true

		//unlock
		prevPred = nil
		for i := highest_locked; i >= 0; i-- {
			if prevPred != prev[i] {
				prev[i].mux.Unlock()
			}
			prevPred = prev[i]

		}

		head.lock.Lock()
		head.n_elements = head.n_elements + 1
		head.lock.Unlock()

		return true
	}

}

func (head *skiplist) Remove(val interface{}) bool {
	/* remove node */

	var nodeToDelete *skiplist_node = nil
	isMarked := false
	top_level := -1

	var prev, next [SKIPLIST_MAX_LEVEL]*skiplist_node

	for {
		// try to find node
		found_level := head.Find(val, prev[:], next[:])

		// if not found or already marked for deletion
		// return false
		if isMarked || (found_level != -1 && CanDelete(next[found_level], found_level)) {
			// not already marked
			if !isMarked {
				// get node
				nodeToDelete = next[found_level]
				top_level = nodeToDelete.top_level
				// lock it
				nodeToDelete.mux.Lock()

				// did some other routine
				// mark it first?
				if nodeToDelete.marked {
					// yes, unlock and abort
					nodeToDelete.mux.Unlock()
					return false
				}

				// no mark it for deletion
				nodeToDelete.marked = true
				isMarked = true
			}

			// now locked

			highest_locked := -1
			var pred, succ *skiplist_node
			var prevPred *skiplist_node = nil

			// validate levels up to top_level
			valid := true
			for level := 0; valid && level <= top_level; level++ {
				pred = prev[level]
				succ = next[level]

				if pred != prevPred {
					pred.mux.Lock()
					highest_locked = level
					prevPred = pred
				}
				valid = !pred.marked && pred.next[level] == succ
			}

			// can't delete try again
			if !valid {

				// unlock to try again
				prevPred = nil
				for i := highest_locked; i >= 0; i-- {
					if prevPred != prev[i] {
						prev[i].mux.Unlock()
					}
					prevPred = prev[i]
				}

				continue
			}
			// actually delete node
			for level := top_level; level >= 0; level-- {
				prev[level].next[level] = nodeToDelete.next[level]
			}

			nodeToDelete.mux.Unlock()

			// cleanup and unlock
			prevPred = nil
			for i := highest_locked; i >= 0; i-- {
				if prevPred != prev[i] {
					prev[i].mux.Unlock()
				}
				prevPred = prev[i]

			}
			// update element count
			head.lock.Lock()
			head.n_elements--
			head.lock.Unlock()

			return true
		} else {
			return false
		}
	}
}

func CanDelete(candidate *skiplist_node, found_level int) bool {
	return candidate.fully_linked && candidate.top_level == found_level && !candidate.marked
}

func Union(skipa *skiplist, skipb *skiplist, inherit_first bool) (merged *skiplist) {
	/* merge two skiplist sets into a new skiplist, keeping the previous two intact.
	If inherit_first is true, use probability and fast_random of skipa else skipb.
	O(N),Not threadsafe */

	// new skiplist
	merged = new(skiplist)

	// max of two levels
	if skipa.n_levels > skipb.n_levels {
		merged.n_levels = skipa.n_levels
	} else {
		merged.n_levels = skipb.n_levels
	}

	// add head node
	merged.head = new(skiplist_node)
	merged.head.fully_linked = true

	// inherit random generation attributes
	// from proper list
	if inherit_first {
		merged.fast_random = skipa.fast_random
		merged.prob = skipa.prob
	} else {
		merged.fast_random = skipb.fast_random
		merged.prob = skipb.prob
	}

	// reset elements
	merged.n_elements = 0

	var aptr, bptr, new_node *skiplist_node
	var prev_elem *interface{} = nil

	// keep last node added in each level
	var prevs [SKIPLIST_MAX_LEVEL]*skiplist_node

	// add head nodes
	for level := merged.n_levels - 1; level >= 0; level-- {
		prevs[level] = merged.head
	}

	// skip head nodes in both origins
	if skipa.head != nil {
		aptr = skipa.head.next[0]
	}
	if skipb.head != nil {
		bptr = skipb.head.next[0]
	}

	/* last level contains all elements sorted
	go through them and add them up to their max level
	while merging */

	var element_to_add *skiplist_node
	/* merge */
	for !(aptr == nil && bptr == nil) {

		/* skip same consecutive elements */
		if prev_elem != nil {
			if aptr != nil && Equals(aptr.value, *prev_elem) {
				aptr = aptr.next[0]
				continue
			} else if bptr != nil && Equals(bptr.value, *prev_elem) {
				bptr = bptr.next[0]
				continue
			}
		}

		// create new node
		new_node = new(skiplist_node)
		new_node.fully_linked = true

		element_to_add = nil

		/* choose if element from first or second list will be added first */
		if (aptr != nil && bptr != nil && Less(aptr.value, bptr.value)) || bptr == nil {
			// keep prev for same check
			prev_elem = &aptr.value
			element_to_add = aptr
			// move first list pointer forward
			aptr = aptr.next[0]
		} else {
			// keep prev for same check
			prev_elem = &bptr.value
			element_to_add = bptr
			// move second list pointer forward
			bptr = bptr.next[0]
		}

		if element_to_add != nil {

			new_node.value = element_to_add.value
			new_node.top_level = element_to_add.top_level

			for level := element_to_add.top_level; level >= 0; level-- {
				prevs[level].next[level] = new_node
				prevs[level] = prevs[level].next[level]
			}

			merged.n_elements++

		}

	}

	return merged

}

func Intersection(skipa *skiplist, skipb *skiplist, inherit_first bool) (intersected *skiplist) {
	/* merge two skiplist sets into a new skiplist, keeping the previous two intact.
	If inherit_first is true, use probability and fast_random of skipa else skipb.
	O(N),Not threadsafe */

	max_level := 0

	// new skiplist
	intersected = new(skiplist)

	// max of two levels
	if skipa.n_levels > skipb.n_levels {
		intersected.n_levels = skipa.n_levels
	} else {
		intersected.n_levels = skipb.n_levels
	}

	// add head node
	intersected.head = new(skiplist_node)
	intersected.head.fully_linked = true

	// inherit random generation attributes
	// from proper list
	if inherit_first {
		intersected.fast_random = skipa.fast_random
		intersected.prob = skipa.prob
	} else {
		intersected.fast_random = skipb.fast_random
		intersected.prob = skipb.prob
	}

	// reset elements
	intersected.n_elements = 0

	var aptr, bptr, new_node *skiplist_node
	var prev_elem *interface{} = nil

	// keep last node added in each level
	var prevs [SKIPLIST_MAX_LEVEL]*skiplist_node

	// add head nodes
	for level := intersected.n_levels - 1; level >= 0; level-- {
		prevs[level] = intersected.head
	}

	// skip head nodes in both origins
	if skipa.head != nil {
		aptr = skipa.head.next[0]
	}
	if skipb.head != nil {
		bptr = skipb.head.next[0]
	}

	/* last level contains all elements sorted
	go through them and add them up to their max level
	while merging */

	/* merge */
	for aptr != nil && bptr != nil {

		/* skip same consecutive elements */
		if prev_elem != nil {
			if aptr != nil && Equals(aptr.value, *prev_elem) {
				aptr = aptr.next[0]
				continue
			} else if bptr != nil && Equals(bptr.value, *prev_elem) {
				bptr = bptr.next[0]
				continue
			}
		}

		/* element in both sets, add */
		if aptr != nil && bptr != nil && Equals(aptr.value, bptr.value) {
			// keep prev for same check
			prev_elem = &aptr.value

			new_node = new(skiplist_node)
			new_node.fully_linked = true

			new_node.value = aptr.value

			// merge by level
			// only levels which have the element in both lists
			// will have the element in the new list
			if aptr.top_level > bptr.top_level {
				new_node.top_level = bptr.top_level
			} else {
				new_node.top_level = aptr.top_level
			}

			if new_node.top_level > max_level {
				max_level = new_node.top_level
			}

			for level := new_node.top_level; level >= 0; level-- {
				prevs[level].next[level] = new_node

				prevs[level] = prevs[level].next[level]
			}

			intersected.n_elements++

			// move first list pointer forward
			aptr = aptr.next[0]
			bptr = bptr.next[0]
		} else if Less(aptr.value, bptr.value) {
			// keep prev for same check
			prev_elem = &aptr.value
			// move second list pointer forward

			aptr = skipa.FindNextLowest(bptr.value)

		} else {
			// keep prev for same check
			prev_elem = &bptr.value
			// move second list pointer forward

			bptr = skipb.FindNextLowest(aptr.value)
		}

	}
	intersected.n_levels = max_level + 1
	return intersected

}

func main() {

}
