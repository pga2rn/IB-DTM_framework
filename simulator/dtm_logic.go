package simulator

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/fwtype"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/vehicle"
	"sync"
)

func (sim *SimulationSession) prepareRSUsForSlot(ctx context.Context, slot uint32) {
	select {
	case <-ctx.Done():
		logutil.GetLogger(PackageName).Fatalf("[prepareRSUsForSlot] context cancel")
	default:
		for i := 0; i < sim.Config.RSUNum; i++ {
			x, y := sim.Config.IndexToCoord(uint32(i))
			r := sim.RSUs[x][y]
			r.InsertSlotsInRing(slot, &fwtype.TrustValueOffsetsPerSlot{})
		}
	}
}

// trust value offsets are stored on each RSU components
func (sim *SimulationSession) genTrustValueOffset(ctx context.Context, slot uint32) {
	logutil.GetLogger(PackageName).Debugf("[genTrustValueOffset] slot %v, epoch %v", slot, slot/sim.Config.SlotsPerEpoch)
	defer logutil.GetLogger(PackageName).Debugf("[genTrustValueOffset] done")

	select {
	case <-ctx.Done():
		logutil.GetLogger(PackageName).Debugf("[genTrustValueOffset] context canceled")
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
					if v.Id != uint32(id) {
						logutil.GetLogger(PackageName).
							Fatalf("[genTrustValueOffset] index and vehicle id mismatches, %v, %v", id, v.Id)
					}

					if v.VehicleStatus != vehicle.Active {
						wg.Done() // job done
						return
					}

					tvo := fwtype.TrustValueOffset{
						VehicleId: v.Id,
						Slot:      slot,
					}

					// adjust trust value weight
					possibility := sim.R.Float32()
					switch {
					case possibility < 0.15:
						tvo.Weight = fwtype.Fatal
					case possibility >= 0.15 && possibility < 0.3:
						tvo.Weight = fwtype.Critical
					default:
						tvo.Weight = fwtype.Routine
					}

					if sim.MisbehaviorVehicleBitMap.Get(int(v.Id)) {
						// the vehicle is assigned to be a bad vehicle
						// md vehicles choose randomly to do evil or not, but do evil more
						// locate to the cross in the map, assign the trust value to cross RSU

						// the idea is that the vehicle will try to do very bad things,
						// or they will behave normally
						flag := sim.R.Float32()
						switch {
						case tvo.Weight == fwtype.Critical && flag < 0.7:
							tvo.TrustValueOffset = -1
						case tvo.Weight == fwtype.Routine && flag < 0.5:
							tvo.TrustValueOffset = 1
						default:
							tvo.TrustValueOffset = -1
						}
					} else {
						// I am a good vehicle!
						// but there is still possible for me to perform some evil when I am malfunction!
						flag := sim.R.Float32()
						switch {
						case flag < 0.10 && tvo.Weight == fwtype.Routine:
							tvo.TrustValueOffset = -1
						default:
							tvo.TrustValueOffset = 1
						}
					}

					// update the value to RSU
					sim.rmu.Lock()
					rsu := sim.RSUs[v.Pos.X][v.Pos.Y]

					// evil type 1: alter the trust value offsets for compromised RSU
					if sim.CompromisedRSUBitMap.Get(int(rsu.Id)) {
						rn := sim.R.Float32()
						// assign altered type
						if rn < 0.8 {
							tvo.AlterType = fwtype.Flipped
						} else {
							tvo.AlterType = fwtype.Dropped
						}
					}
					rsu.GetSlotInRing(slot).Store(v.Id, &tvo)

					sim.rmu.Unlock() // finish trust value offsets injection

					wg.Done() // job done
				} // select
			}(id, v) // go routine
		} // for loop

		// wait for all jobs to be done
		wg.Wait()
	} // select
}

// execute compromised RSU logic here
func (sim *SimulationSession) forgeTrustValueOffsets(ctx context.Context, slot uint32) {
	logutil.GetLogger(PackageName).Debugf("[forgeTrustValueOffsets] slot %v", slot)
	defer logutil.GetLogger(PackageName).Debugf("[forgeTrustValueOffsets] slot %v, done", slot)

	select {
	case <-ctx.Done():
		logutil.GetLogger(PackageName).Fatalf("[forgeTrustValueOffsets] slot %v, context canceled", slot)
	default:
		wg := sync.WaitGroup{}

		// spawn a go routine for each RSU
		for i := 0; i < sim.Config.RSUNum; i++ {

			x, y := sim.Config.IndexToCoord(uint32(i))
			rsu := sim.RSUs[x][y]
			rsu.ManagedVehicles = sim.Map.GetCross(rsu.Pos).GetVehicleNum()

			wg.Add(1)
			go func() {
				// execute compromised RSU evil logics
				if sim.CompromisedRSUBitMap.Get(int(rsu.Id)) {
					rn, target := sim.R.Float32(), 0
					// if managed vehicles num is too small
					// the compromised RSU will not do evils to hide themselves
					if rsu.ManagedVehicles < sim.ActiveVehiclesNum/sim.Config.RSUNum {
						wg.Done()
						return
					} else {
						target = rsu.ManagedVehicles
					}

					switch {
					case rn < 0.6:
						target = target / 2
					case rn >= 0.6 && rn < 0.9:
						target = target * 4 / 5
					default:
						target = sim.R.RandIntRange(target, target*2)
					}

					for i := 0; i < target; i++ {
						vid := uint32(sim.R.RandIntRange(0, sim.Config.VehicleNumMax))

						tvo := &fwtype.TrustValueOffset{
							AlterType: fwtype.Forged,
							VehicleId: vid,
						}

						// randomly rate the vehicle
						rn := sim.R.Float32()
						switch {
						case rn < 0.7: // portion of good vehicles are larger, so rate lower it
							tvo.TrustValueOffset = -1
						default:
							tvo.TrustValueOffset = 1
						}

						rn = sim.R.Float32()
						switch {
						case rn < 0.2:
							tvo.Weight = fwtype.Fatal
						case rn >= 0.2 && rn < 0.5:
							tvo.Weight = fwtype.Critical
						default:
							tvo.Weight = fwtype.Routine
						}

						// store the forged data into RSU storage area
						rsu.GetSlotInRing(slot).Store(vid, tvo)
					}
				} // if is evil RSU
				wg.Done()
			}() // go routine
		} // iterate RSU for loop
		wg.Wait()
	} // select
}
