package simulator

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/dtmtype"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/vehicle"
)

func (sim *SimulationSession) executeDTMLogicPerSlot(ctx context.Context, slot uint64) {
	logutil.LoggerList["core"].Debugf("[executeDTMLogicPerSlot] entering ..")
	select {
	case <-ctx.Done():
		logutil.LoggerList["core"].Debugf("[executeDTMLogicPerSlot] context canceled")
		return
	default:
		// init the data structure for every RSU to store tvos for the slot
		sim.prepareRSUsForSlot(ctx, slot)
		// generate and dispatch trust value offsets to every RSUs
		sim.genTrustValueOffset(ctx, slot)
	}
}

func (sim *SimulationSession) prepareRSUsForSlot(ctx context.Context, slot uint64){
	logutil.LoggerList["core"].Debugf("[prepareRSUsForSlot] slot %v", slot)
	select {
	case <-ctx.Done():
		return
	default:
		for x:= 0; x < sim.Config.XLen;x++{
			for y := 0; y<sim.Config.YLen;y++{
				sim.rmu.Lock()
				rsu := sim.RSUs[x][y]
				sim.rmu.Unlock()
				rsu.InsertSlotsInRing(slot, &dtmtype.TrustValueOffsetsPerSlot{})
			}
		}
	}
}


// trust value offsets are stored on each RSU components
func (sim *SimulationSession) genTrustValueOffset(ctx context.Context, slot uint64) {
	logutil.LoggerList["core"].Debugf("[genTrustValueOffset] slot %v, epoch %v", slot, slot/sim.Config.SlotsPerEpoch)
	defer logutil.LoggerList["core"].Debugf("[genTrustValueOffset] done")


	select {
	case <-ctx.Done():
		logutil.LoggerList["core"].Debugf("[genTrustValueOffset] context canceled")
		return
	default:
		// iterate every vehicles, and then generate trust value offsets for it
		for id, v := range sim.Vehicles {
			go func(id int, v *vehicle.Vehicle) {
				if v.Id != uint64(id) {
					logutil.LoggerList["core"].
						Fatalf("[genTrustValueOffset] index and vehicle id mismatches, %v, %v", id, v.Id)
				}

				if v.VehicleStatus != vehicle.Active {
					return
				}

				tvo := dtmtype.TrustValueOffset{
					VehicleId: v.Id,
					Slot:      slot,
				}

				if sim.MisbehaviorVehicleBitMap.Get(int(v.Id)) {
					// the vehicle is assigned to be a bad vehicle
					// md vehicles choose randomly to do evil or not, but do evil more
					// locate to the cross in the map, assign the trust value to cross RSU
					flag := sim.R.Float32()
					switch {
					// 10% possibility to no be evil
					case flag < 0.1:
						tvo.TrustValueOffset = 1
					default:
						tvo.TrustValueOffset = -1
					}
				} else {
					// I am a good vehicle!
					tvo.TrustValueOffset = 1
				}

				// adjust trust value weight
				possibility := sim.R.Float32()
				switch {
				case possibility < 1-dtmtype.Fatal:
					tvo.Weight = dtmtype.Fatal
				case possibility < 1-dtmtype.Critical && possibility > 1-dtmtype.Fatal:
					tvo.Weight = dtmtype.Critical
				default:
					tvo.Weight = dtmtype.Routine
				}

				// update the value to RSU
				// update each slot
				sim.rmu.Lock()
				sim.RSUs[v.Pos.X][v.Pos.Y].GetSlotsInRing(slot).Store(v.Id, &tvo)
				sim.rmu.Unlock()
			}(id, v) // go routine
		}// for loop
	}
}
