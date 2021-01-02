package rsu

import (
	"github.com/pga2rn/ib-dtm_framework/shared/dtmtype"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/pair"
	"sync"
)

// RSU will storage N epochs trust value offsets data
type RSU struct {
	// unique id of an RSU, index in the sim-session object
	Id uint32

	// pos
	Pos pair.Position

	// trust value offsets storage
	ring    *dtmtype.TrustValueOffsetsPerSlotRing
	ringLen int

	// managed zone info is stored at the cross object
	// the most recent slot's managed vehicles num
	ManagedVehicles int

	// for dtm logic use
	nextSlotForUpload uint32 // the slot that available for uploading trust value offset
	uploadMu          sync.Mutex
}

// type of evil
// only for reference, see dtmtype/trust_value_offset.go
const (
	FlipTrustValueOffset = iota
	DropPositiveTrustValueOffset
	ForgeTrustValueOffset
)

// RSU constructor
func InitRSU(id uint32, pos pair.Position, ringLen int) *RSU {
	return &RSU{
		Id:       id,
		uploadMu: sync.Mutex{},
		Pos:      pos,
		ringLen:  ringLen,
		ring:     dtmtype.InitRing(ringLen),
	}
}

func (rsu *RSU) InsertSlotsInRing(slot uint32, element *dtmtype.TrustValueOffsetsPerSlot) {
	//logutil.LoggerList["dtm"].Debugf("[InsertSlotsInRing] RSU %v, slot %v", rsu.Id, slot)
	baseSlot, curSlot := rsu.ring.GetProperties()

	if curSlot+1 != slot && slot != 0 {
		logutil.LoggerList["dtm"].Fatalf("[InserSlotsInRing] rsu %v, curSlot %v, slot %v", rsu.Id, curSlot, slot)
		return
	}

	if curSlot >= uint32(rsu.ringLen) { // ring is full
		baseSlot += 1
	}
	curSlot = slot
	rsu.ring.SetElement(element, baseSlot, curSlot)
}

func (rsu *RSU) GetSlotInRing(slot uint32) *dtmtype.TrustValueOffsetsPerSlot {
	rin, rinMu := rsu.ring.GetRing()
	baseSlot, curSlot := rsu.ring.GetProperties()

	if slot < baseSlot || slot > curSlot {
		return nil
	}

	rinMu.Lock()
	res := rin.Move(-int(curSlot - slot)).Value.(*dtmtype.TrustValueOffsetsPerSlot)
	rinMu.Unlock()

	return res
}

func (rsu *RSU) GetNextUploadSlot() uint32 {
	rsu.uploadMu.Lock()
	res := rsu.nextSlotForUpload
	rsu.uploadMu.Unlock()
	return res
}

// input is the latest uploaded slot
func (rsu *RSU) SetNextUploadSlot(slot uint32) {
	if slot < rsu.nextSlotForUpload {
		return
	}
	rsu.uploadMu.Lock()
	if slot == 0 {
		rsu.nextSlotForUpload = 0
	} else {
		rsu.nextSlotForUpload = slot + 1
	}
	rsu.uploadMu.Unlock()
}

func (rsu *RSU) GetRingInformation() (baseSlot, currentSlot uint32) {
	return rsu.ring.GetProperties()
}
