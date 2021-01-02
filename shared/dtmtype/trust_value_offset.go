package dtmtype

import (
	"container/ring"
	"sync"
)

// trust value starts range from -1 ~ 1
// starts from 0

type TrustValueOffset struct {
	VehicleId        uint32
	Slot             uint32
	TrustValueOffset float32
	Weight           float32
	// for compromisedRSU
	AlterType int
}

// alter type const
const (
	Flipped = iota
	Dropped
	Forged
)

// we use sync.map for thread safe
type TrustValueOffsetsPerSlot = sync.Map // map[<vehicleId>uint32]*TrustValueOffset
//type TrustValueOffsetsPerEpoch = sync.Map // map[<slot>uint32]*TrustValueOffsetsPerSlot

// trust value offset weight
const (
	Routine  = 0.15
	Critical = 0.5
	Fatal    = 0.9
)

// sync.map can be used directly without extra initializing
type TrustValueOffsetsPerSlotRing struct {
	mu                    sync.Mutex
	r                     *ring.Ring // *TrustValueOffsetsPerSlot
	baseSlot, currentSlot uint32     // ring base slot
}

func InitRing(len int) *TrustValueOffsetsPerSlotRing {
	return &TrustValueOffsetsPerSlotRing{
		mu:          sync.Mutex{},
		r:           ring.New(len),
		baseSlot:    0,
		currentSlot: 0,
	}
}

func (r *TrustValueOffsetsPerSlotRing) SetElement(element *TrustValueOffsetsPerSlot, base, current uint32) {
	r.mu.Lock()
	rin := r.r.Next()
	rin.Value = element
	// update current head
	r.r = rin
	r.baseSlot, r.currentSlot = base, current
	r.mu.Unlock()
}

func (r *TrustValueOffsetsPerSlotRing) GetRing() (*ring.Ring, *sync.Mutex) {
	return r.r, &r.mu
}

func (r *TrustValueOffsetsPerSlotRing) GetProperties() (baseSlot uint32, currentSlot uint32) {
	return r.baseSlot, r.currentSlot
}
