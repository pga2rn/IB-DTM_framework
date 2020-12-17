package core

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/dtmutils"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/simulator/vehicle"
)

// trust value offsets are stored on each RSU components
func (sim *SimulationSession) genTrustValueOffset(ctx context.Context, slot uint64) {
	select {
	case <- ctx.Done():
		logutil.LoggerList["core"].Debugf("[genTrustValueOffset] context canceled")
		return
	default:
		// iterate every vehicles, and then generate trust value offsets for it
		for id, v := range sim.Vehicles {
			if v.Id != uint64(id){
				logutil.LoggerList["core"].
					Fatalf("[genTrustValueOffset] index and vehicle id mismatches")
			}

			if v.VehicleStatus != vehicle.Active{
				continue
			}

			slotIndex := int(slot % sim.Config.SlotsPerEpoch)
			tvo := dtmutils.TrustValueOffset{
				VehicleId: v.Id,
				Slot: slot,
			}

			if sim.MisbehaviorVehicleBitMap.Get(int(v.Id)) {
				// the vehicle is assigned to be a bad vehicle
				// md vehicles choose randomly to do evil or not, but do evil more
				// locate to the cross in the map, assign the trust value to cross RSU
				flag := sim.R.Float32()
				switch {
				// 20% possibility to no be evil
				case flag < 0.2:
					tvo.TrustValueOffset = 1
				default:
					tvo.TrustValueOffset = -1
				}
			} else {
				// I am a good vehicle!
				tvo.TrustValueOffset = 1
			}

			// update the value to RSU
			// update each slot
			sim.RSUs[v.Pos.X][v.Pos.Y].
				TrustValueOffsetPerSlot[slotIndex][v.Id] = &tvo
		}
	}
}