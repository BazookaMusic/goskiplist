package goskiplist

import (
	"fmt"
)

/*Height get max Skiplist level */
func (list *Skiplist) Height() int {
	/* current max level */
	defer list.lock.RUnlock()
	list.lock.RLock()
	return list.nLevels
}

/*Len get number of inserted unique elements */
func (list *Skiplist) Len() int {
	/* current dataAmount of inserted elements */
	defer list.lock.RUnlock()
	list.lock.RLock()
	return list.nElements
}

/* set max levels,
if levels > SkiplistMaxLevel, set to SkiplistMaxLevel,
if levels <= 0, set to 1
Threadsafe*/
func (list *Skiplist) setMaxLevels(levels int) {
	defer list.lock.Unlock()
	list.lock.Lock()
	list.maxLevels = min(max(1, levels), SkiplistMaxLevel)
}

/* set prob,
if prob > 1, set to SkiplistMaxLevel,
if prob < MinProb, set to MinProb
Threadsafe*/
func (list *Skiplist) setProb(prob float64) {
	defer list.lock.Unlock()
	list.lock.Lock()
	list.prob = minF(maxF(MinProb, prob), 1.0)
}

/* set fastRandom,
Threadsafe*/
func (list *Skiplist) setFastRandom(isSet bool) {
	defer list.lock.Unlock()
	list.lock.Lock()
	list.fastRandom = isSet
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
		maxLevels = SkiplistMaxLevel
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

/*ToSortedArray : Return sorted array of inserted Skiplist items, not threadsafe */
func (list *Skiplist) ToSortedArray() []SkiplistItem {
	/* make a sorted array out of the Skiplist
	   returns the lowest level               */
	arr := make([]SkiplistItem, list.nElements, list.nElements)
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
func (list *Skiplist) findNextLowest(val SkiplistItem) (node *skiplistNode) {
	// could be modified by inserts
	list.lock.RLock()
	level := list.nLevels - 1
	list.lock.RUnlock()
	// much faster than starting at max

	pred := list.head

	var curr *skiplistNode

	// traverse vertically
	for ; level >= 0; level-- {
		// horizontally
		curr = pred.next[level]
		for curr != nil && curr.value.Less(val) {
			pred = curr
			curr = pred.next[level]
		}

		// next of where it should be
		if curr != nil && curr.value.Equals(val) {
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
func (list *Skiplist) Find(val SkiplistItem, prev, next []*skiplistNode) (foundLevel int) {

	// could be modified by inserts
	list.lock.RLock()
	level := list.nLevels - 1
	list.lock.RUnlock()
	// much faster than starting at max

	pred := list.head
	foundLevel = -1
	var curr *skiplistNode

	// traverse vertically
	for ; level >= 0; level-- {
		// horizontally
		curr = pred.next[level]
		for curr != nil && curr.value.Less(val) {
			pred = curr
			curr = pred.next[level]
		}

		// next of where it should be
		if curr != nil && curr.value.Equals(val) && foundLevel == -1 {
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
func (list *Skiplist) Contains(val SkiplistItem) bool {

	list.lock.RLock()
	level := list.nLevels - 1
	list.lock.RUnlock()

	pred := list.head
	var curr *skiplistNode
	// vertically
	for ; level >= 0; level-- {
		// horizontally
		curr = pred.next[level]
		for curr != nil && curr.value.Less(val) {
			pred = curr
			curr = pred.next[level]
		}
		//found something or have to go down

		// is the next element what I seek
		if curr != nil && curr.value.Equals(val) {
			node := curr
			return node.fullyLinked && !node.marked
		}
	}
	// not found
	return false
}

/*Insert : Insert node with value v to Skiplist. Returns true on success,false on failure to insert.
Thread safe. */
func (list *Skiplist) Insert(v SkiplistItem) bool {
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
func (list *Skiplist) Remove(val SkiplistItem) bool {
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
func (list *Skiplist) Union(skipa, skipb *Skiplist) *Skiplist {

	// can't have less max levels than its current levels
	list.maxLevels = max(list.maxLevels, max(skipa.maxLevels, skipb.maxLevels))

	// add head node
	list.head = new(skiplistNode)
	list.head.fullyLinked = true

	// reset elements
	list.nElements = 0

	union(list, skipa, skipb, true)
	return list
}

/*UnionSimple Merge two Skiplist sets into a new Skiplist, keeping the previous two intact.
The new Skiplist levels will be the merged levels of the two skiplists.

New levels are not generated. The # of max levels of the new Skiplist is readjusted
to allow merge.

New skiplist parameters set to defaults of:

list.prob = 0.5,

list.fastRandom = true,

list.maxLevels = SkiplistMaxLevel


Returns new Skiplist.

O(N),Not threadsafe */
func UnionSimple(skipa, skipb *Skiplist) *Skiplist {
	list := new(Skiplist)
	list.prob = 0.5
	list.fastRandom = true
	list.maxLevels = SkiplistMaxLevel

	// max of two levels
	list.nLevels = max(skipa.nLevels, skipb.nLevels)

	// add head node
	list.head = new(skiplistNode)
	list.head.fullyLinked = true

	// readjust max levels to make union possible
	list.maxLevels = max(skipa.nLevels, list.maxLevels) // can't have less levels than its current
	list.maxLevels = max(list.maxLevels, skipb.nLevels)

	// reset elements
	list.nElements = 0

	union(list, skipa, skipb, false)
	return list

}

/* actual implementation */
func union(list, skipa, skipb *Skiplist, newProb bool) *Skiplist {

	if newProb {
		list.nLevels = 0
	} else {
		list.nLevels = max(skipa.nLevels, skipb.nLevels)
	}

	var aptr, bptr, newNode *skiplistNode
	var prevElem *SkiplistItem

	// keep last node added in each level
	var prevs [SkiplistMaxLevel]*skiplistNode

	// add head nodes
	for level := list.maxLevels - 1; level >= 0; level-- {
		prevs[level] = list.head
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
			if aptr != nil && aptr.value.Equals(*prevElem) {
				aptr = aptr.next[0]
				continue
			} else if bptr != nil && bptr.value.Equals(*prevElem) {
				bptr = bptr.next[0]
				continue
			}
		}

		// create new node
		newNode = new(skiplistNode)
		newNode.fullyLinked = true

		elementToAdd = nil

		/* choose if element from first or second list will be added first */
		if (aptr != nil && bptr != nil && aptr.value.Less(bptr.value)) || bptr == nil {
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
				newNode.topLevel = coinTosses(list.prob, list.maxLevels, list.fastRandom) - 1
				list.nLevels = max(newNode.topLevel+1, list.nLevels)
			}

			for level := newNode.topLevel; level >= 0; level-- {
				prevs[level].next[level] = newNode
				prevs[level] = prevs[level].next[level]
			}

			list.nElements++

		}

	}

	return list

}

/*Intersection Intersect two Skiplist sets into a new Skiplist, keeping the previous two intact.
The new Skiplist parameters will define the structure of the new Skiplist,
meaning that the top levels of insertion of each node will be generated again
O(N),Not threadsafe */
func (list *Skiplist) Intersection(skipa, skipb *Skiplist) *Skiplist {

	// max of two levels
	list.nLevels = max(skipa.nLevels, skipb.nLevels)
	// can't have less max levels than its current levels
	list.maxLevels = max(list.maxLevels, list.nLevels)

	// add head node
	list.head = new(skiplistNode)
	list.head.fullyLinked = true

	// reset elements
	list.nElements = 0

	intersection(list, skipa, skipb, true)
	return list

}

/*IntersectionSimple Intersect two Skiplist sets into a new Skiplist, keeping the previous two intact.
The new Skiplist levels will be the intersection of the other two skiplists' levels,
new insertion levels will not be generated. (Faster than Intersect)

New skiplist parameters set to defaults of:

list.prob = 0.5,

list.fastRandom = true,

list.maxLevels = SkiplistMaxLevel

Not threadsafe */
func IntersectionSimple(skipa, skipb *Skiplist) *Skiplist {

	list := new(Skiplist)

	// reset parameters
	list.prob = 0.5
	list.fastRandom = true
	list.maxLevels = SkiplistMaxLevel

	// max of two levels
	list.nLevels = max(skipa.nLevels, skipb.nLevels)

	// add head node
	list.head = new(skiplistNode)
	list.head.fullyLinked = true

	// reset elements
	list.nElements = 0

	intersection(list, skipa, skipb, false)
	return list

}

func intersection(intersected, skipa, skipb *Skiplist, newProb bool) *Skiplist {
	/* merge two Skiplist sets into a new Skiplist, keeping the previous two intact.
	If inherit_first is true, use probability and fastRandom of skipa else skipb.
	O(N),Not threadsafe */

	// new Skiplist top level
	maxLevel := 0

	// intersection will have at most as many levels
	// as the shortest Skiplist
	intersected.nLevels = min(skipa.nLevels, skipb.nLevels)

	// add head node
	intersected.head = new(skiplistNode)
	intersected.head.fullyLinked = true

	// reset elements
	intersected.nElements = 0

	var aptr, bptr, newNode *skiplistNode
	var prevElem *SkiplistItem

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
			if aptr.value.Equals(*prevElem) {
				aptr = aptr.next[0]
				continue
			} else if bptr.value.Equals(*prevElem) {
				bptr = bptr.next[0]
				continue
			}
		}

		/* element in both sets, add */
		if aptr != nil && bptr != nil && aptr.value.Equals(bptr.value) {
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
				newNode.topLevel = min(aptr.topLevel, bptr.topLevel)
			}

			// update new total Skiplist maximum level
			maxLevel = max(newNode.topLevel, maxLevel)

			for level := newNode.topLevel; level >= 0; level-- {
				prevs[level].next[level] = newNode

				prevs[level] = prevs[level].next[level]
			}

			intersected.nElements++

			// move list pointers forward
			aptr = aptr.next[0]
			bptr = bptr.next[0]
		} else if aptr.value.Less(bptr.value) {
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
