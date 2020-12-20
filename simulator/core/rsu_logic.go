package core

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/dtmutil"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/randutil"
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
	"github.com/pga2rn/ib-dtm_framework/simulator/dtm"
)

func (sim *SimulationSession) InitRSU() bool {
	for x := range sim.RSUs {
		// init every RSU data structure
		for y := range sim.RSUs[x] {
			r := sim.RSUs[x][y]

			r.Id = uint64(y)*uint64(sim.Config.XLen) + uint64(x)
			r.Epoch = 0
			r.Slot = 0

			// uploading tracker
			r.NextSlotForUpload = 0

			// init the data structure of trust value offset storage
			r.TrustValueOffsetPerSlot =
				make([]map[uint64]*dtmutil.TrustValueOffset, sim.Config.SlotsPerEpoch)
			for i := range r.TrustValueOffsetPerSlot {
				// init map structure for every slot
				r.TrustValueOffsetPerSlot[i] = make(map[uint64]*dtmutil.TrustValueOffset)
			}

			// not yet connected with external RSU module
			r.ExternalRSUModuleInitFlag = false
		}
	}

	return true
}

//
func (sim *SimulationSession) resetRSUAtCheckpoint(ctx context.Context, slot uint64) {
	if slot%sim.Config.SlotsPerEpoch != 0 {
		logutil.LoggerList["core"].Warnf("[resetRSUAtCheckpoint] being called at non-checkpoint slot, abort")
		return
	}

	// reset the rsu trust value offset storage
	for x := range sim.RSUs {
		for y := range sim.RSUs[x] {
			rsu := sim.RSUs[x][y]
			sim.resetRSUTrustValueOffsetStorage(rsu)
		}
	}
}

// reset RSU data fields at the end of epoch
func (sim *SimulationSession) resetRSUTrustValueOffsetStorage(r *dtm.RSU) {
	r.TrustValueOffsetPerSlot =
		make([]map[uint64]*dtmutil.TrustValueOffset, sim.Config.SlotsPerEpoch)
	for i := range r.TrustValueOffsetPerSlot {
		// init map structure for every slot
		r.TrustValueOffsetPerSlot[i] = make(map[uint64]*dtmutil.TrustValueOffset)
	}
}

// do rsu should do
func (sim *SimulationSession) executeRSULogic(ctx context.Context, slot uint64) {
	select {
	case <-ctx.Done():
		return
	default:
		// update epoch and slot
		for x := range sim.RSUs {
			for y := range sim.RSUs[x] {
				// sync epoch and slot
				r := sim.RSUs[x][y]
				r.Slot, r.Epoch =
					timeutil.SlotsSinceGenesis(sim.Config.Genesis),
					timeutil.EpochsSinceGenesis(sim.Config.Genesis)

				if r.Slot != slot {
					logutil.LoggerList["core"].Debugf("[executeRSULogic] async with slot!")
				}
			}
		}
	}
}

// helper function for processEpoch to assign compromisedRSU
func (sim *SimulationSession) initAssignCompromisedRSU(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	default:
		count := 0
		sim.CompromisedRSUPortion = randutil.RandFloatRange(
			sim.R,
			sim.Config.CompromisedRSUPortionMin,
			sim.Config.CompromisedRSUPortionMax,
		)
		target := int(float32(sim.Config.RSUNum) * sim.CompromisedRSUPortion)

		for count < target {
			index := randutil.RandIntRange(sim.R, 0, sim.Config.RSUNum)
			if !sim.CompromisedRSUBitMap.Get(index) {
				sim.CompromisedRSUBitMap.Set(index, true)
				count++
			}
		}
	}
}
