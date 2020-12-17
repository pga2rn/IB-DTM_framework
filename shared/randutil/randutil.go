// Package randutil is a simple wrapper for go math/rand lib
package randutil

import (
	"math/rand"
)

func InitRand(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

// return an Int within the range[start, stop)
func RandIntRange(r *rand.Rand, start, stop int) int {
	// convert stop-start to float32, multiplied by a portion factor,
	// and then rounded it down to int
	increment := int(r.Float32() * float32(r.Intn(stop - start)))
	return start + increment
}

// return an float32 within the range[start, stop)
func RandFloatRange(r *rand.Rand, start, stop float32) float32 {
	increment := r.Float32() * (stop - start)
	return start + increment
}
