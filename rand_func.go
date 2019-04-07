package goskiplist

import (
	"math/rand"
)

const mask = ((1 << SkiplistMaxLevel) - 1)

// FastImplementation = true, one call to rand, 0.5 probability
// FastImplementation = false, consecutive calls to rand, variable probability
const FastImplementation = true

func coinTosses(prob float64, maxLevels int, fast bool) (counter int) {

	counter = 1
	// very fast with probability 0.5
	// only one call to rand
	// find first zero in random float bit representation
	// geometric distribution
	if fast {

		resMask := rand.Uint64() & mask

		// find first zero in float representation
		for ; resMask&1 == 0; resMask >>= 1 {
			counter++
		}

		return counter

	}

	// supports probability
	// slower
	res := rand.Float64()
	for res < prob {
		res = rand.Float64()
		counter++
	}
	return counter

}
