package simulator

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/rsu"
)

func (sim *SimulationSession) InitRSUs() bool {
	for x := range sim.RSUs {
		// init every RSU data structure
		for y := range sim.RSUs[x] {
			r := rsu.InitRSU(
				uint32(sim.CoordToIndex(x, y)),
				x, y,
				sim.Config.RingLength,
			)

			// uploading tracker
			r.SetNextUploadSlot(0)
			sim.RSUs[x][y] = r
		}
	}

	return true
}

// helper function for processEpoch to assign compromisedRSU
func (sim *SimulationSession) initAssignCompromisedRSU(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	default:
		count := 0
		sim.CompromisedRSUPortion = sim.R.RandFloatRange(
			sim.Config.CompromisedRSUPortionMin,
			sim.Config.CompromisedRSUPortionMax,
		)
		target := int(float32(sim.Config.RSUNum) * sim.CompromisedRSUPortion)

		for count < target {
			index := sim.R.RandIntRange(0, sim.Config.RSUNum)
			if !sim.CompromisedRSUBitMap.Get(index) {
				sim.CompromisedRSUBitMap.Set(index, true)
				count++
			}
		}
	}
}
