package dtm

import (
	"github.com/pga2rn/ib-dtm_framework/shared/dtmutils"
	"github.com/pga2rn/ib-dtm_framework/simulator/vehicle"
)

type RSU struct {
	// unique id of an RSU, index in the sim-session object
	Id uint64

	// for sync
	Epoch uint64
	Slot uint64

	// if connected with external RSU module
	ExternalRSUModuleInitFlag bool

	// management zone
	// id of vehicle
	Vehicle map[uint64]*vehicle.Vehicle
	// map of trust value offset per slot
	TrustValueOffsetPerSlot []map[uint64]dtmutils.TrustValueOffset
	// to indicate the rsu to be compromised or not, aligned with TrustValueOffsetPerSlot
	CompromisedFlag	bool

	// for external rsu module
	NextSlotForUpload uint64 // the slot that available for uploading trust value offset
}

// provide trust value offsets for external RSU module
//func (rsu *RSU) ProvideTrustValueOffsets