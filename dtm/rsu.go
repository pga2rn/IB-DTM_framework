package dtm

import (
	"github.com/pga2rn/ib-dtm_framework/shared/dtmutil"
)

type position struct {
	X int
	Y int
}

type RSU struct {
	// unique id of an RSU, index in the sim-session object
	Id uint64

	// pos
	Pos position

	// for sync
	Epoch uint64
	Slot  uint64

	// if connected with external RSU module
	ExternalRSUModuleInitFlag bool

	// management zone
	// id of vehicle (check it on Map.cross)
	// Vehicle map[uint64]*vehicle.Vehicle

	// map of trust value offset per slot
	// this is for external RSU modules
	TrustValueOffsetPerSlot []map[uint64]*dtmutil.TrustValueOffset
	// to indicate the rsu to be compromised or not, aligned with TrustValueOffsetPerSlot
	// DEPRECATED: move to session
	//CompromisedFlag	bool

	// for external rsu module
	NextSlotForUpload uint64 // the slot that available for uploading trust value offset
}

// provide trust value offsets for external RSU module
//func (rsu *RSU) ProvideTrustValueOffsets
