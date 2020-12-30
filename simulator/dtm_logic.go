package simulator

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/dtmtype"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/vehicle"
	"sync"
)

func (sim *SimulationSession) executeDTMLogicPerSlot(ctx context.Context, slot uint64) {
	logutil.LoggerList["simulator"].Debugf("[executeDTMLogicPerSlot] entering ..")
	select {
	case <-ctx.Done():
		logutil.LoggerList["simulator"].Debugf("[executeDTMLogicPerSlot] context canceled")
		return
	default:
		// init the data structure for every RSU to store tvos for the slot
		sim.prepareRSUsForSlot(ctx, slot)
		// generate and dispatch trust value offsets to every RSUs
		sim.genTrustValueOffset(ctx, slot)
	}
}

func (sim *SimulationSession) prepareRSUsForSlot(ctx context.Context, slot uint64) {
	logutil.LoggerList["simulator"].Debugf("[prepareRSUsForSlot] slot %v", slot)
	select {
	case <-ctx.Done():
		logutil.LoggerList["simulator"].Fatalf("[prepareRSUsForSlot] context cancel")
		return
	default:
		for x := 0; x < sim.Config.XLen; x++ {
			for y := 0; y < sim.Config.YLen; y++ {
				sim.rmu.Lock()
				r := sim.RSUs[x][y]
				sim.rmu.Unlock()
				r.InsertSlotsInRing(slot, &dtmtype.TrustValueOffsetsPerSlot{})
			}
		}
	}
}

// trust value offsets are stored on each RSU components
func (sim *SimulationSession) genTrustValueOffset(ctx context.Context, slot uint64) {
	logutil.LoggerList["simulator"].Debugf("[genTrustValueOffset] slot %v, epoch %v", slot, slot/sim.Config.SlotsPerEpoch)
	defer logutil.LoggerList["simulator"].Debugf("[genTrustValueOffset] done")

	select {
	case <-ctx.Done():
		logutil.LoggerList["simulator"].Debugf("[genTrustValueOffset] context canceled")
		return
	default:
		wg := sync.WaitGroup{}

		// iterate every vehicles, and then generate trust value offsets for it
		for id, v := range sim.Vehicles {
			wg.Add(1)

			go func(id int, v *vehicle.Vehicle) {
				select {
				case <-ctx.Done():
					wg.Done() // job done
					return
				default:
					if v.Id != uint64(id) {
						logutil.LoggerList["simulator"].
							Fatalf("[genTrustValueOffset] index and vehicle id mismatches, %v, %v", id, v.Id)
					}

					if v.VehicleStatus != vehicle.Active {
						wg.Done() // job done
						return
					}

					tvo := dtmtype.TrustValueOffset{
						VehicleId: v.Id,
						Slot:      slot,
					}

					// adjust trust value weight
					possibility := sim.R.Float32()
					switch {
					case possibility < 0.1:
						tvo.Weight = dtmtype.Fatal
					case possibility >= 0.1 && possibility < 0.25:
						tvo.Weight = dtmtype.Critical
					default:
						tvo.Weight = dtmtype.Routine
					}

					if sim.MisbehaviorVehicleBitMap.Get(int(v.Id)) {
						// the vehicle is assigned to be a bad vehicle
						// md vehicles choose randomly to do evil or not, but do evil more
						// locate to the cross in the map, assign the trust value to cross RSU

						// the idea is that the vehicle will try to do very bad things,
						// or they will behave normally
						flag := sim.R.Float32()
						switch {
						case tvo.Weight == dtmtype.Routine:
							tvo.TrustValueOffset = 1
						case tvo.Weight == dtmtype.Critical && flag < 0.3:
							tvo.TrustValueOffset = -1
						default:
							tvo.TrustValueOffset = -1
						}
					} else {
						// I am a good vehicle!
						tvo.TrustValueOffset = 1
					}

					// update the value to RSU
					// update each slot
					sim.rmu.Lock()
					sim.RSUs[v.Pos.X][v.Pos.Y].GetSlotInRing(slot).Store(v.Id, &tvo)
					sim.rmu.Unlock()

					wg.Done() // job done
				} // select
			}(id, v) // go routine
		} // for loop

		// wait for all jobs to be done
		wg.Wait()
	} // select
}
