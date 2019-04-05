This is a thread safe implementation of a generic skiplist in Go. It is a direct implementation of the simple 
optimistic lazy Skiplist algorithm from the paper by Maurice Herlihy,Yossi Lev,Victor Luchangco and Nir Shavit.

[Link to paper](http://people.csail.mit.edu/shanir/publications/LazySkipList.pdf "the paper")

The implementation includes tests and the repository is connected with Travis for continuous integration.

Example usage:
```golang
    // initialise random number generator
        rand.Seed(time.Now().UTC().UnixNano())

    // allocate skiplist struct
    head := new(Skiplist)


    /* Initialise skiplist parameters */

    // the first parameter is the probability of
    // the bernoulli trials to choose the maximum list
    // level

    // the second parameter is the max amount of levels
    // which will contain items. 
    // SkiplistMaxLevel is the max amount allowed by the
    // implementation (compile constant)

    // the third parameter, if set to true,
    // enables an optimised algorithm to
    // generate the random levels but
    // with a set probability of 0.5


    head.InitSkiplist(0.5, SkiplistMaxLevel, false)

    /* thread-safe insert,remove, contains */
    dataAmount := 100

    // insert

    for index := 0; index < dataAmount; index++ {
		if !head.Insert(interface{}(index)) {
			// insertion failed, item already inserted
		}
	}

    // contains
    for index := 0; index < dataAmount; index++ {
		if !head.Contains(interface{}(index)) {
			// item not in skiplist
		}
	}

    //remove
     for index := 0; index < dataAmount; index++ {
		if !head.Remove(interface{}(index)) {
			//  item not in skiplist
		}
    }
    
    /* will not corrupt the structure */
    for index := 0; index < dataAmount; index++ {
        go head.Insert(interface{}(index))
        go head.Contains(interface{}(index))
        go head.Remove(interface{}(index))
        
	}
    


```

