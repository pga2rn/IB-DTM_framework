package dtm

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/simulator/core"
)

type RSU struct {
	// unique id of an RSU, index in the sim-session object
	Id uint64
	Session *core.SimulationSession

	// for sync
	TimeSync core.Beacon

	// management zone
	// id of vehicle
	Vehicle []uint64
	// map of trust value offset per slot
	TrustValueOffsetPerSlot []*map[uint64]TrustValueOffset
}

// init RSU
func InitRSU(sim *core.SimulationSession) *RSU {
	rsu := &RSU{}
	rsu.TimeSync.Genesis = sim.Config.Genesis
	return rsu
}

// called by the simulator
// the simulator provide trust value to the RSU
func (rsu *RSU) ProcessSlot (
	ctx context.Context,
	b core.Beacon,
	trustOffsetList *map[uint64]TrustValueOffset) error {

		// sync the timesync using genesis time and current time

	// check if the slot and epoch is correct
	switch {
	case rsu.TimeSync.Genesis != b.Genesis:
		return nil
	case rsu.TimeSync.Epoch != b.Epoch:
		// out of sync too much
		return nil
	case rsu.TimeSync.Slot - b.Slot >= rsu.Session.Config.OutOfSyncTolerant:
		// TODO: fix out of sync judgement here
		return nil
	}

	// push newly generate trust value to the list
	rsu.TrustValueOffsetPerSlot[b.Slot % b.Epoch] = trustOffsetList
	return nil
}

// at the start of the epoch, calculate the previous
func (rsu *RSU) ProcessEpoch(ctx context.Context) error {
	if rsu.TimeSync.Slot % rsu.TimeSync.Epoch != 0 {
		return nil
	}

	if rsu.TimeSync.Epoch - rsu.Session.Config.FinalizedDelay <= 0 {
		return nil
	}

	// here we call the function in package trust_value to calculate the trust value for every vehicles

	// then we push the
	// HEAD here
}