package rsu

import (
	"github.com/pga2rn/ib-dtm_framework/shared/fwtype"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
)

var PackageName = "rsu"

// RSU will storage N epochs trust value offsets data
type RSU struct {
	// unique id of an RSU, index in the sim-session object
	Id uint32

	// pos
	Pos fwtype.Position

	// trust value offsets storage
	ring    *fwtype.TrustValueOffsetsPerSlotRing
	ringLen int

	// managed zone info is stored at the cross object
	// the most recent slot's managed vehicles num
	ManagedVehicles         int
	ManagedVehiclesPerEpoch int
}

// type of evil
// only for reference, see dtmtype/trust_value_offset.go
const (
	FlipTrustValueOffset = iota
	DropPositiveTrustValueOffset
	ForgeTrustValueOffset
)

// RSU constructor
func InitRSU(id uint32, pos fwtype.Position, ringLen int) *RSU {
	return &RSU{
		Id:      id,
		Pos:     pos,
		ringLen: ringLen,
		ring:    fwtype.InitRing(ringLen),
	}
}

func (rsu *RSU) InsertSlotsInRing(slot uint32, element *fwtype.TrustValueOffsetsPerSlot) {
	//logutil.GetLogger(PackageName).Debugf("[InsertSlotsInRing] RSU %v, slot %v", rsu.Id, slot)
	baseSlot, curSlot := rsu.ring.GetProperties()

	if curSlot+1 != slot && slot != 0 {
		logutil.GetLogger(PackageName).Fatalf("[InserSlotsInRing] rsu %v, curSlot %v, slot %v", rsu.Id, curSlot, slot)
		return
	}

	if curSlot >= uint32(rsu.ringLen) { // ring is full
		baseSlot += 1
	}
	curSlot = slot
	rsu.ring.SetElement(element, baseSlot, curSlot)
}

func (rsu *RSU) GetSlotInRing(slot uint32) *fwtype.TrustValueOffsetsPerSlot {
	return rsu.ring.GetElementForSlot(slot)
}
func (rsu *RSU) GetRingInformation() (baseSlot, currentSlot uint32) {
	return rsu.ring.GetProperties()
}
