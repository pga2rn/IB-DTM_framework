// Package syncutil implements low level atomic operations over float32/float64
package syncutil

import (
	"math"
	"sync/atomic"
	"unsafe"
)

func atomicLoadFloat64(x *float64) float64 {
	return math.Float64frombits(atomic.LoadUint64((*uint64)(unsafe.Pointer(x))))
}

func atomicLoadFloat32(x *float32) float32 {
	return math.Float32frombits(atomic.LoadUint32((*uint32)(unsafe.Pointer(x))))
}

func atomicStoreFloat64(x *float64, val float64) {
	atomic.StoreUint64((*uint64)(unsafe.Pointer(x)), math.Float64bits(val))
}

func atomicStoreFloat32(x *float32, val float32) {
	atomic.StoreUint32((*uint32)(unsafe.Pointer(x)), math.Float32bits(val))
}