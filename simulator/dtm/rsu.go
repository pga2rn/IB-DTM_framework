package dtm

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/simulator/core"
	"github.com/pga2rn/ib-dtm_framework/simulator/sim-map"
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
	TrustValueOffsetPerSlot []map[uint64]TrustValueOffset
}

// init RSU
func InitRSU(sim *core.SimulationSession) *RSU {
	rsu := &RSU{}
	rsu.TimeSync.Genesis = sim.Config.Genesis
	return rsu
}

// called by the simulator
// the simulator provide trust value
func (rsu *RSU) ProcessSlot (
	ctx context.Context,
	offsetlist *[]TrustValueOffset) error {


}

//
func (rsu *RSU) ProcessEpoch(ctx context.Context) error {

}