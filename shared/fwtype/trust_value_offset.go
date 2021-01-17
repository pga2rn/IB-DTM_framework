package fwtype

import (
	"container/ring"
	"sync"
)

var PackageName = "fwtype"

// trust value starts range from -1 ~ 1
// starts from 0

type TrustValueOffset struct {
	VehicleId        uint32
	Slot             uint32
	TrustValueOffset float32
	Weight           float32
	// for compromisedRSU
	AlterType TrustValueOffsetAlertedType
}

type TrustValueOffsetAlertedType = uint32

// alter type const
const (
	Flipped = 233
	Dropped = 234
	Forged  = 235
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
	mu                    sync.RWMutex
	r                     *ring.Ring // *TrustValueOffsetsPerSlot
	baseSlot, currentSlot uint32     // ring base slot
}

func InitRing(len int) *TrustValueOffsetsPerSlotRing {
	return &TrustValueOffsetsPerSlotRing{
		mu:          sync.RWMutex{},
		r:           ring.New(len),
		baseSlot:    0,
		currentSlot: 0,
	}
}

func (r *TrustValueOffsetsPerSlotRing) SetElement(element *TrustValueOffsetsPerSlot, base, current uint32) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rin := r.r.Next()
	rin.Value = element
	// update current head
	r.r = rin
	r.baseSlot, r.currentSlot = base, current
}

func (r *TrustValueOffsetsPerSlotRing) GetElementForSlot(slot uint32) *TrustValueOffsetsPerSlot {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, curSlot := r.GetProperties()
	return r.r.Move(-int(curSlot - slot)).Value.(*TrustValueOffsetsPerSlot)
}

func (r *TrustValueOffsetsPerSlotRing) GetRing() (*ring.Ring, *sync.RWMutex) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.r, &r.mu
}

func (r *TrustValueOffsetsPerSlotRing) GetProperties() (baseSlot uint32, currentSlot uint32) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.baseSlot, r.currentSlot
}
