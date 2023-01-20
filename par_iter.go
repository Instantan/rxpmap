package rxpmap

import (
	"runtime"
	"sync"
)

const PAR_FOREACH_JUNK_SIZE = 100
const MAX_GO_ROUTINES_MULTIPLICATOR = 3

// ParForeach takes a slice and applies the given func to each
// element. It automatically distributes the slice into multiple go routines based on
// length of the given slice.
// The functions blocks till all ops are completed
func parForeach[T any](s []T, fn func(T)) {
	if len(s) < PAR_FOREACH_JUNK_SIZE {
		for i := range s {
			fn(s[i])
		}
		return
	}
	numCpu := runtime.NumCPU()
	lenS := len(s)
	chunkSize := ((lenS + numCpu - 1) / numCpu) * MAX_GO_ROUTINES_MULTIPLICATOR
	wg := sync.WaitGroup{}
	for i := 0; i < lenS; i += chunkSize {
		end := i + chunkSize
		if end > lenS {
			end = lenS
		}
		go foreachSync(s[i:end], fn, &wg)
	}
	wg.Wait()
}

func foreachSync[T any](s []T, fn func(T), wg *sync.WaitGroup) {
	for i := range s {
		fn(s[i])
	}
	wg.Done()
}
