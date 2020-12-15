package dtm

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/dtmutils"
	"github.com/pga2rn/ib-dtm_framework/simulator/core"
	"github.com/pga2rn/ib-dtm_framework/simulator/sim-map"
	"github.com/pga2rn/ib-dtm_framework/shared/dtmutils"
)

type RSU struct {
	// unique id of an RSU, index in the simsession object
	Id uint64

	// for sync
	TimeSync core.Beacon

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
func InitRSU(sim *core.SimulationSession) *RSU {
	rsu := &RSU{}
	rsu.TimeSync.Genesis = sim.Config.Genesis
	return rsu
}

// provide trust value offsets for external RSU module
//func (rsu *RSU) ProvideTrustValueOffsets

// called by the simulator
// accept the trust value offsets
func (rsu *RSU) ProcessSlot (
	ctx context.Context,
	slot uint64,
	compromisedFlag bool,
	offsetlist *[]dtmutils.TrustValueOffset) error {
	// update the

}