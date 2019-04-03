package main

import "fmt"

// just prints the Skiplist
func skiplistDebug(a *Skiplist) {
	level := a.nLevels - 1
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
