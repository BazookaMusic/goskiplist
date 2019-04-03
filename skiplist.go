package main

import (
	"fmt"
)

/*Height get max Skiplist level */
func (list *Skiplist) Height() int {
	/* current max level */
	defer list.lock.Unlock()
	list.lock.Lock()
	return list.nLevels
}

/*Len get number of inserted unique elements */
func (list *Skiplist) Len() int {
	/* current dataAmount of inserted elements */
	defer list.lock.Unlock()
	list.lock.Lock()
	return list.nElements
}

/*InitSkiplist :
prob : Probability of bernoulli trials to find level of insertion.
maxLevels: max level of insertion
fastRandom: true -> use optimised random level generation with set probability 0.5 (fast),
false -> use bernoulli trials with consecutive calls to random (slower but variable probability) */
func (list *Skiplist) InitSkiplist(prob float64, maxLevels int, fastRandom bool) {

	if prob < 0 {
		prob = 0.5
		fmt.Println("Init: Probability given less than zero, set to 0.5 instead")
	}

	if maxLevels > SkiplistMaxLevel {
		fmt.Println("Init: Max level given more than supported dataAmount of",
			SkiplistMaxLevel, " setting to ", SkiplistMaxLevel, "instead")
	}

	list.nLevels = 1
	list.prob = prob
	list.maxLevels = maxLevels
	list.fastRandom = fastRandom

	newHead := new(skiplistNode)
	newHead.fullyLinked = true
	newHead.marked = false

	list.nElements = 0

	list.head = newHead
}

/*ToSortedArray : Return sorted array of inserted Skiplist items */
func (list *Skiplist) ToSortedArray() []interface{} {
	/* make a sorted array out of the Skiplist
	   returns the lowest level               */
	arr := make([]interface{}, list.nElements, list.nElements)
	counter := 0
	for currentNode := list.head.next[0]; currentNode != nil; currentNode = currentNode.next[0] {
		arr[counter] = currentNode.value
		counter++
	}

	return arr

}

/*findNextLowest : Find where the element should be
and return it's successor on the first level.
Returns the element or
-1 when not found */
func (list *Skiplist) findNextLowest(val interface{}) (node *skiplistNode) {
	// could be modified by inserts
	list.lock.Lock()
	level := list.nLevels - 1
	list.lock.Unlock()
	// much faster than starting at max

	pred := list.head

	var curr *skiplistNode

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

/*Find : Find where the node with value val should be in the Skiplist,
return the first level where it was found and the
next and previous elements for every level.
Returns the first level where it was found or
-1 when not found */
func (list *Skiplist) Find(val interface{}, prev, next []*skiplistNode) (foundLevel int) {

	// could be modified by inserts
	list.lock.Lock()
	level := list.nLevels - 1
	list.lock.Unlock()
	// much faster than starting at max

	pred := list.head
	foundLevel = -1
	var curr *skiplistNode

	// traverse vertically
	for ; level >= 0; level-- {
		// horizontally
		curr = pred.next[level]
		for curr != nil && Less(curr.value, val) {
			pred = curr
			curr = pred.next[level]
		}

		// next of where it should be
		if curr != nil && Equals(curr.value, val) && foundLevel == -1 {
			foundLevel = level
		}

		// previous of where the item should be
		prev[level] = pred
		next[level] = curr

	}

	return foundLevel
}

/*Contains : Return true if node with value val exists in Skiplist,
else false. */
func (list *Skiplist) Contains(val interface{}) bool {

	list.lock.Lock()
	level := list.nLevels - 1
	list.lock.Unlock()

	pred := list.head
	var curr *skiplistNode
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
			return node.fullyLinked && !node.marked
		}
	}
	// not found
	return false
}

/*Insert : Insert node with value v to Skiplist. Returns true on success,false on failure to insert.
Thread safe. */
func (list *Skiplist) Insert(v interface{}) bool {
	// insert element

	// highest level of insertion
	// the list.fast property should not be modified after init
	topLevel := coinTosses(list.prob, list.maxLevels, list.fastRandom)

	// check if list must become taller
	list.lock.Lock()
	if topLevel > list.nLevels {
		list.nLevels = topLevel
	}
	list.lock.Unlock()

	// buffers to store prev and next pointers
	var prev, next []*skiplistNode
	prev = make([]*skiplistNode, SkiplistMaxLevel)
	next = make([]*skiplistNode, SkiplistMaxLevel)

	for {

		// find insertion point and previous and next nodes
		foundLevel := list.Find(v, prev, next)

		// already in Skiplist
		if foundLevel != -1 {

			// should be the node with value v
			nodeFound := next[foundLevel]
			// if node is not set for removal
			if !nodeFound.marked {
				// wait until stable
				for !nodeFound.fullyLinked {
				}
				//don't insert
				return false
			}
			// try again
			continue

		}
		// highest level locked
		highestLocked := -1
		var pred, succ *skiplistNode
		var prevPred *skiplistNode

		valid := true

		// validate that new node can be added
		// by checking previous and next nodes
		for level := 0; valid && level < topLevel; level++ {

			pred = prev[level]
			succ = next[level]

			// avoid locking same node twice
			// if two or more levels
			// connected to same node
			if pred != prevPred {
				pred.mux.Lock()

				highestLocked = level
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
			for i := highestLocked; i >= 0; i-- {
				if prevPred != prev[i] {
					prev[i].mux.Unlock()
				}
				prevPred = prev[i]

			}
			// restart attempt
			continue
		}

		// try to add new node
		newNode := new(skiplistNode)
		newNode.value = v
		newNode.topLevel = topLevel - 1
		newNode.marked = false

		for level := 0; level < topLevel; level++ {

			newNode.next[level] = next[level]
			prev[level].next[level] = newNode
		}
		// new node is ok
		newNode.fullyLinked = true

		//unlock
		prevPred = nil
		for i := highestLocked; i >= 0; i-- {
			if prevPred != prev[i] {
				prev[i].mux.Unlock()
			}
			prevPred = prev[i]

		}

		list.lock.Lock()
		list.nElements = list.nElements + 1
		list.lock.Unlock()

		return true
	}

}

/*Remove : Remove node with value val from Skiplist, if ite exists. Returns true on success,
false on not found or failure to remove. Thread safe. */
func (list *Skiplist) Remove(val interface{}) bool {
	/* remove node */

	var nodeToDelete *skiplistNode
	isMarked := false
	topLevel := -1

	var prev, next [SkiplistMaxLevel]*skiplistNode

	for {
		// try to find node
		foundLevel := list.Find(val, prev[:], next[:])

		// if not found or already marked for deletion
		// return false
		if isMarked || (foundLevel != -1 && canDelete(next[foundLevel], foundLevel)) {
			// not already marked
			if !isMarked {
				// get node
				nodeToDelete = next[foundLevel]
				topLevel = nodeToDelete.topLevel
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

			highestLocked := -1
			var pred, succ *skiplistNode
			var prevPred *skiplistNode

			// validate levels up to topLevel
			valid := true
			for level := 0; valid && level <= topLevel; level++ {
				pred = prev[level]
				succ = next[level]

				if pred != prevPred {
					pred.mux.Lock()
					highestLocked = level
					prevPred = pred
				}
				valid = !pred.marked && pred.next[level] == succ
			}

			// can't delete try again
			if !valid {

				// unlock to try again
				prevPred = nil
				for i := highestLocked; i >= 0; i-- {
					if prevPred != prev[i] {
						prev[i].mux.Unlock()
					}
					prevPred = prev[i]
				}

				continue
			}
			// actually delete node
			for level := topLevel; level >= 0; level-- {
				prev[level].next[level] = nodeToDelete.next[level]
			}

			nodeToDelete.mux.Unlock()

			// cleanup and unlock
			prevPred = nil
			for i := highestLocked; i >= 0; i-- {
				if prevPred != prev[i] {
					prev[i].mux.Unlock()
				}
				prevPred = prev[i]

			}
			// update element count
			list.lock.Lock()
			list.nElements--
			list.lock.Unlock()

			return true
		}

		return false

	}
}

// helper
func canDelete(candidate *skiplistNode, foundLevel int) bool {
	return candidate.fullyLinked && candidate.topLevel == foundLevel && !candidate.marked
}

/*Union Merge two Skiplist sets into a new Skiplist, keeping the previous two intact.
The new Skiplist parameters will define the structure of the new Skiplist,
meaning that the top levels of each node will be generated again
O(N),Not threadsafe */
func Union(newSkiplist *Skiplist, skipa *Skiplist, skipb *Skiplist) *Skiplist {

	// can't have less max levels than its current levels
	newSkiplist.maxLevels = Max(newSkiplist.maxLevels, newSkiplist.nLevels)

	// add head node
	newSkiplist.head = new(skiplistNode)
	newSkiplist.head.fullyLinked = true

	// reset elements
	newSkiplist.nElements = 0

	union(newSkiplist, skipa, skipb, true)
	return newSkiplist
}

/*UnionSimple Merge two Skiplist sets into a new Skiplist, keeping the previous two intact.
The new Skiplist levels will be the merged levels of the two skiplists.

New levels are not generated. The # of max levels of the new Skiplist is readjusted
to allow merge.


Returns new Skiplist.

O(N),Not threadsafe */
func UnionSimple(newSkiplist, skipa, skipb *Skiplist) *Skiplist {

	// max of two levels
	newSkiplist.nLevels = Max(skipa.nLevels, skipb.nLevels)

	// add head node
	newSkiplist.head = new(skiplistNode)
	newSkiplist.head.fullyLinked = true

	// readjust max levels to make union possible
	newSkiplist.maxLevels = Max(skipa.nLevels, newSkiplist.maxLevels) // can't have less levels than its current
	newSkiplist.maxLevels = Max(newSkiplist.maxLevels, skipb.nLevels)

	// reset elements
	newSkiplist.nElements = 0

	union(newSkiplist, skipa, skipb, false)
	return newSkiplist

}

/* actual implementation */
func union(merged, skipa, skipb *Skiplist, newProb bool) *Skiplist {

	if newProb {
		merged.nLevels = 0
	} else {
		merged.nLevels = Max(skipa.nLevels, skipb.nLevels)
	}

	var aptr, bptr, newNode *skiplistNode
	var prevElem *interface{}

	// keep last node added in each level
	var prevs [SkiplistMaxLevel]*skiplistNode

	// add head nodes
	for level := merged.maxLevels - 1; level >= 0; level-- {
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

	var elementToAdd *skiplistNode
	/* merge */
	for !(aptr == nil && bptr == nil) {

		/* skip same consecutive elements */
		if prevElem != nil {
			if aptr != nil && Equals(aptr.value, *prevElem) {
				aptr = aptr.next[0]
				continue
			} else if bptr != nil && Equals(bptr.value, *prevElem) {
				bptr = bptr.next[0]
				continue
			}
		}

		// create new node
		newNode = new(skiplistNode)
		newNode.fullyLinked = true

		elementToAdd = nil

		/* choose if element from first or second list will be added first */
		if (aptr != nil && bptr != nil && Less(aptr.value, bptr.value)) || bptr == nil {
			// keep prev for same check
			prevElem = &aptr.value
			elementToAdd = aptr
			// move first list pointer forward
			aptr = aptr.next[0]
		} else {
			// keep prev for same check
			prevElem = &bptr.value
			elementToAdd = bptr
			// move second list pointer forward
			bptr = bptr.next[0]
		}

		if elementToAdd != nil {
			newNode.value = elementToAdd.value

			// keep previous structure or
			//  generate new Skiplist of given probability
			if !newProb {
				newNode.topLevel = elementToAdd.topLevel
			} else {
				newNode.topLevel = coinTosses(merged.prob, merged.maxLevels, merged.fastRandom) - 1
				merged.nLevels = Max(newNode.topLevel+1, merged.nLevels)
			}

			for level := newNode.topLevel; level >= 0; level-- {
				prevs[level].next[level] = newNode
				prevs[level] = prevs[level].next[level]
			}

			merged.nElements++

		}

	}

	return merged

}

/*Intersection Intersect two Skiplist sets into a new Skiplist, keeping the previous two intact.
The new Skiplist parameters will define the structure of the new Skiplist,
meaning that the top levels of insertion of each node will be generated again
O(N),Not threadsafe */
func Intersection(newSkiplist, skipa, skipb *Skiplist) *Skiplist {

	// max of two levels
	newSkiplist.nLevels = Max(skipa.nLevels, skipb.nLevels)
	// can't have less max levels than its current levels
	newSkiplist.maxLevels = Max(newSkiplist.maxLevels, newSkiplist.nLevels)

	// add head node
	newSkiplist.head = new(skiplistNode)
	newSkiplist.head.fullyLinked = true

	// reset elements
	newSkiplist.nElements = 0

	intersection(newSkiplist, skipa, skipb, true)
	return newSkiplist

}

/*IntersectionSimple Intersect two Skiplist sets into a new Skiplist, keeping the previous two intact.
The new Skiplist levels will be the intersection of the other two skiplists' levels,
new insertion levels will not be generated. (Faster than Intersect)
Not threadsafe */
func IntersectionSimple(newSkiplist, skipa, skipb *Skiplist) *Skiplist {

	// max of two levels
	newSkiplist.nLevels = Max(skipa.nLevels, skipb.nLevels)
	// can't have less max levels than its current levels
	newSkiplist.maxLevels = Max(newSkiplist.maxLevels, newSkiplist.nLevels)

	// add head node
	newSkiplist.head = new(skiplistNode)
	newSkiplist.head.fullyLinked = true

	// reset elements
	newSkiplist.nElements = 0

	intersection(newSkiplist, skipa, skipb, false)
	return newSkiplist

}

func intersection(intersected, skipa, skipb *Skiplist, newProb bool) *Skiplist {
	/* merge two Skiplist sets into a new Skiplist, keeping the previous two intact.
	If inherit_first is true, use probability and fastRandom of skipa else skipb.
	O(N),Not threadsafe */

	// new Skiplist top level
	maxLevel := 0

	// intersection will have at most as many levels
	// as the shortest Skiplist
	intersected.nLevels = Min(skipa.nLevels, skipb.nLevels)

	// add head node
	intersected.head = new(skiplistNode)
	intersected.head.fullyLinked = true

	// reset elements
	intersected.nElements = 0

	var aptr, bptr, newNode *skiplistNode
	var prevElem *interface{}

	// keep last node added in each level
	var prevs [SkiplistMaxLevel]*skiplistNode

	// add head nodes
	for level := intersected.maxLevels - 1; level >= 0; level-- {
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
		if prevElem != nil {
			if Equals(aptr.value, *prevElem) {
				aptr = aptr.next[0]
				continue
			} else if Equals(bptr.value, *prevElem) {
				bptr = bptr.next[0]
				continue
			}
		}

		/* element in both sets, add */
		if aptr != nil && bptr != nil && Equals(aptr.value, bptr.value) {
			// keep prev for same check
			prevElem = &aptr.value

			newNode = new(skiplistNode)
			newNode.fullyLinked = true

			newNode.value = aptr.value

			// merge by level
			// only levels which have the element in both lists
			// will have the element in the new list
			if newProb {
				newNode.topLevel = coinTosses(intersected.prob, intersected.maxLevels, intersected.fastRandom) - 1
			} else {
				newNode.topLevel = Min(aptr.topLevel, bptr.topLevel)
			}

			// update new total Skiplist maximum level
			maxLevel = Max(newNode.topLevel, maxLevel)

			for level := newNode.topLevel; level >= 0; level-- {
				prevs[level].next[level] = newNode

				prevs[level] = prevs[level].next[level]
			}

			intersected.nElements++

			// move list pointers forward
			aptr = aptr.next[0]
			bptr = bptr.next[0]
		} else if Less(aptr.value, bptr.value) {
			// keep prev for same check
			prevElem = &aptr.value
			// move second list pointer forward

			aptr = skipa.findNextLowest(bptr.value)

		} else {
			// keep prev for same check
			prevElem = &bptr.value
			// move second list pointer forward

			bptr = skipb.findNextLowest(aptr.value)
		}

	}
	intersected.nLevels = maxLevel + 1
	return intersected

}

func main() {

}
