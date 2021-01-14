// Package randutil is a simple wrapper for go math/rand lib
package randutil

// math/rand is not thread safe, so we wrap the rand with mutex

import (
	"math/rand"
	"sync"
)

type RandUtil struct {
	r  *rand.Rand
	mu sync.Mutex
}

func InitRand(seed int64) *RandUtil {
	return &RandUtil{
		rand.New(rand.NewSource(seed)),
		sync.Mutex{},
	}
}

// the sum of ws should be 1
func (randUtil *RandUtil) Possibility(ws ...float32) int {
	rn := randUtil.Float32()

	count := float32(0)
	for i, w := range ws {
		count += w
		if count > rn {
			return i
		}
	}

	return -1
}

// wrap the basic rand functions with mutex
func (randUtil *RandUtil) Float32() float32 {
	randUtil.mu.Lock()
	num := randUtil.r.Float32()
	randUtil.mu.Unlock()
	return num
}

func (randUtil *RandUtil) Float64() float64 {
	randUtil.mu.Lock()
	num := randUtil.r.Float64()
	randUtil.mu.Unlock()
	return num
}

func (randUtil *RandUtil) Intn(n int) int {
	randUtil.mu.Lock()
	num := randUtil.r.Intn(n)
	randUtil.mu.Unlock()
	return num
}

func (randUtil *RandUtil) Perm(n int) []int {
	randUtil.mu.Lock()
	res := randUtil.r.Perm(n)
	randUtil.mu.Unlock()
	return res
}

func (randUtil *RandUtil) PermUint32(n int) []uint32 {
	randUtil.mu.Lock()
	restmp := randUtil.r.Perm(n)
	randUtil.mu.Unlock()

	res := make([]uint32, n)
	for i := range restmp {
		res[i] = uint32(restmp[i])
	}
	return res
}

// advanced random function
// return an Int within the range[start, stop)
func (randUtil *RandUtil) RandIntRange(start, stop int) int {
	// convert stop-start to float32, multiplied by a portion factor,
	// and then rounded it down to int
	return start + randUtil.Intn(stop-start)
}

// return an float32 within the range[start, stop)
func (randUtil *RandUtil) RandFloatRange(start, stop float32) float32 {
	return start + randUtil.Float32()*(stop-start)
}
