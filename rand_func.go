package main

import (
	"math/rand"
	"time"
)

const Mask = ((1 << SKIPLIST_MAX_LEVEL) - 1)
const FAST_IMPLEMENTATION = true

func coin_tosses(prob float64, max_levels int, fast bool) (counter int) {
	t1 := time.Now()

	counter = 1
	// very fast with probability 0.5
	// only one call to rand
	if fast {

		res_mask := rand.Uint64() & Mask

		// find first zero in float representation
		for ; res_mask&1 == 0; res_mask >>= 1 {
			counter++
		}

	} else {
		// supports probability
		// slower
		res := rand.Float64()
		for res < prob {
			res = rand.Float64()
			counter++
		}

	}

	randtime += time.Now().Sub(t1)
	return counter

}
