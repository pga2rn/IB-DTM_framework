package dtm

import (
	"github.com/pga2rn/ib-dtm_framework/shared/dtmutils"
)

type RSU struct {
	// unique id of an RSU, index in the sim-session object
	Id uint64

	// for sync
	Epoch uint64
	Slot uint64

	// management zone
	// id of vehicle
	Vehicle []uint64
	// map of trust value offset per slot
	TrustValueOffsetPerSlot []map[uint64]dtmutils.TrustValueOffset
	// to indicate the rsu to be compromised or not, aligned with TrustValueOffsetPerSlot
	CompromisedFlagPerSlot []bool

	// for external rsu module
	NextSlotForUpload uint64 // the slot that available for uploading trust value offset
}

// init simulated RSU
func InitRSU() *RSU {
	rsu := &RSU{}
	return rsu
}

// provide trust value offsets for external RSU module
//func (rsu *RSU) ProvideTrustValueOffsets