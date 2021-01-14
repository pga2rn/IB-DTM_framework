package simulator

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/rsu"
	"github.com/pga2rn/ib-dtm_framework/shared/fwtype"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
)

func (sim *SimulationSession) InitRSUs() bool {
	for x := range sim.RSUs {
		// init every RSU data structure
		for y := range sim.RSUs[x] {
			r := rsu.InitRSU(
				uint32(sim.Config.CoordToIndex(x, y)),
				fwtype.Position{x, y},
				sim.Config.RingLength,
			)

			sim.RSUs[x][y] = r
		}
	}

	return true
}

// helper function for processEpoch to assign compromisedRSU
func (sim *SimulationSession) initAssignCompromisedRSU(ctx context.Context) {
	select {
	case <-ctx.Done():
		logutil.LoggerList["simulator"].Fatalf("[initAssignMisbehaveVehicle] context canceled")
		return
	default:
		sim.CompromisedRSUPortion = sim.Config.CompromisedRSUPortion

		target := int(float32(sim.Config.RSUNum) * sim.CompromisedRSUPortion)

		for i := 0; i < target; i++ {
			index := sim.R.RandIntRange(0, sim.Config.RSUNum)
			if !sim.CompromisedRSUBitMap.Get(index) {
				sim.CompromisedRSUBitMap.Set(index, true)
			}
		}
	}
}

// Evil type 1: alter the existed trust value offsets
// the altered trust value offsets will finally being altered when trust values are being calculated
// dive into the slot
// move to genTrustValueOffsets

// evil type 2
// Evil type 2, forge trust value offsets
// store the updated tvo back to RSU
// RSU will try to make more vehicles being treated as misbehaving,
// to over thrown the dtm itself
