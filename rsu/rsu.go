package rsu

import (
	"github.com/pga2rn/ib-dtm_framework/shared/dtmtype"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"sync"
)

type position struct {
	X int
	Y int
}

// RSU will storage N epochs trust value offsets data
type RSU struct {
	// unique id of an RSU, index in the sim-session object
	Id uint64

	// pos
	Pos position

	// trust value offsets storage
	ring    *dtmtype.TrustValueOffsetsPerSlotRing
	ringLen int

	// for dtm logic use
	nextSlotForUpload uint64 // the slot that available for uploading trust value offset
	uploadMu          sync.Mutex
}

// type of evil
const (
	FlipTrustValueOffset = iota
	DropPositiveTrustValueOffset
	ForgeTrustValueOffset
)

var RSUEvilsType = []int{FlipTrustValueOffset, DropPositiveTrustValueOffset}

// RSU constructor
func InitRSU(id uint64, x, y int, ringLen int) *RSU {
	return &RSU{
		Id:       id,
		uploadMu: sync.Mutex{},
		Pos:      position{x, y},
		ringLen:  ringLen,
		ring:     dtmtype.InitRing(ringLen),
	}
}

func (rsu *RSU) InsertSlotsInRing(slot uint64, element *dtmtype.TrustValueOffsetsPerSlot) {
	//logutil.LoggerList["dtm"].Debugf("[InsertSlotsInRing] RSU %v, slot %v", rsu.Id, slot)
	baseSlot, curSlot := rsu.ring.GetProperties()

	if curSlot+1 != slot && slot != 0 {
		logutil.LoggerList["dtm"].Fatalf("[InserSlotsInRing] rsu %v, curSlot %v, slot %v", rsu.Id, curSlot, slot)
		return
	}

	if curSlot >= uint64(rsu.ringLen) { // ring is full
		baseSlot += 1
	}
	curSlot = slot
	rsu.ring.SetElement(element, baseSlot, curSlot)
}

func (rsu *RSU) GetSlotInRing(slot uint64) *dtmtype.TrustValueOffsetsPerSlot {
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

func (rsu *RSU) GetNextUploadSlot() uint64 {
	rsu.uploadMu.Lock()
	res := rsu.nextSlotForUpload
	rsu.uploadMu.Unlock()
	return res
}

// input is the latest uploaded slot
func (rsu *RSU) SetNextUploadSlot(slot uint64) {
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

func (rsu *RSU) GetRingInformation() (baseSlot, currentSlot uint64) {
	return rsu.ring.GetProperties()
}
