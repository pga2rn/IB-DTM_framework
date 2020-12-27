package core

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/dtm"
	"github.com/pga2rn/ib-dtm_framework/shared/dtmtype"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/timefactor"
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
	"github.com/pga2rn/ib-dtm_framework/vehicle"
	"sync"
)

func (sim *SimulationSession) executeDTMLogicPerSlot(ctx context.Context, slot uint64) {
	logutil.LoggerList["core"].Debugf("[executeDTMLogicPerSlot] entering ..")
	select {
	case <-ctx.Done():
		logutil.LoggerList["core"].Debugf("[executeDTMLogicPerSlot] context canceled")
		return
	default:
		// generate and dispatch trust value offsets to every RSUs
		sim.genTrustValueOffset(ctx, slot)
		// update rsu status (maybe a kind of redundant)
		sim.executeRSULogic(ctx, slot)
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
			if v.Id != uint64(id) {
				logutil.LoggerList["core"].
					Fatalf("[genTrustValueOffset] index and vehicle id mismatches, %v, %v", id, v.Id)
			}

			if v.VehicleStatus != vehicle.Active {
				continue
			}

			slotIndex := int(slot % sim.Config.SlotsPerEpoch)
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
			case possibility < 1-dtmtype.Crital && possibility > 1-dtmtype.Fatal:
				tvo.Weight = dtmtype.Crital
			default:
				tvo.Weight = dtmtype.Rountine
			}

			// update the value to RSU
			// update each slot
			sim.RSUs[v.Pos.X][v.Pos.Y].
				TrustValueOffsetPerSlot[slotIndex][v.Id] = &tvo
		}
	}
}

func (sim *SimulationSession) genTrustValue(ctx context.Context, slot uint64) {
	logutil.LoggerList["core"].Debugf("[genTrustValue] start to process for epoch %v", slot/sim.Config.SlotsPerEpoch)
	defer logutil.LoggerList["core"].
		Debugf("[genTrustValue] epoch %v, slot %v done", slot/sim.Config.SlotsPerEpoch, slot)

	select {
	case <-ctx.Done():
		return
	default:
		if slot != timeutil.SlotsSinceGenesis(sim.Config.Genesis) {
			logutil.LoggerList["core"].
				Warnf("[genTrustValue] mismatch slot index! potential async")
		}
		if slot%sim.Config.SlotsPerEpoch != 0 {
			logutil.LoggerList["core"].Fatalf("[genTrustValue] call func at non checkpoint slot, abort")
		}

		// init a data structure to store the trust value
		trustValueRecord :=
			dtmtype.InitTrustValueStorageObject(slot / sim.Config.SlotsPerEpoch)

		wg := sync.WaitGroup{}

		// iterate all RSU
		// set deadline
		epochCtx, cancel := context.WithDeadline(ctx, timeutil.NextEpochTime(sim.Config.Genesis, slot))
		for x := range sim.RSUs {
			for y := range sim.RSUs[x] {
				rsu := sim.RSUs[x][y]

				// use go routines to collect every RSU's data
				// add one worker to wait group
				wg.Add(1)
				go func() {
					select {
					case <-epochCtx.Done():
						logutil.LoggerList["core"].Fatalf("[genTrustValue] times up for collecting RSU data at the end of epoch, abort")
						return
					default:
						// RSU: for every slots
						for slotIndex := 0; slotIndex < int(sim.Config.SlotsPerEpoch); slotIndex++ {
							// get the slot
							slotInstance := rsu.TrustValueOffsetPerSlot[slotIndex]
							// dive into the slot
							for vid, tvo := range slotInstance {
								if vid != tvo.VehicleId {
									logutil.LoggerList["core"].
										Warnf("[genTrustValue] mismatch vid! %v in vehicle and %v in tvo", vid, tvo.VehicleId)
									continue // ignore invalid trust value offset record
								}

								//
								tunedTrustValueOffset := sim.genTrustValueHelper(rsu, tvo.TrustValueOffset, slot)
								if op, ok := trustValueRecord.TrustValueList.LoadOrStore(vid, tunedTrustValueOffset); ok {
									// ok means there is already value stored in place
									// the existed value is loaded to variable op
									trustValueRecord.TrustValueList.Store(vid, op.(float32)+tunedTrustValueOffset)
								}
							}
						}
					} // select
					wg.Done() // job done,
				}() // go routine
			}
		}

		// wait for all work to finish their job
		wg.Wait()
		// after all the workers finish their job, cancel the context
		cancel()

		// TODO: realize background tracking services to keep records of trust value
		sim.TrustValueList = trustValueRecord.TrustValueList
	}
}

// return tuned trust value offset!(may be or may not be compromised!)
func (sim *SimulationSession) genTrustValueHelper(rsu *dtm.RSU, tvo float32, slot uint64) float32 {
	timeFactor := timefactor.GetTimeFactor(
		sim.Config.TimeFactorType,
		sim.Config.Genesis,
		timeutil.SlotStartTime(sim.Config.Genesis, slot),
		timeutil.NextEpochTime(sim.Config.Genesis, slot),
	)
	res := float32(timeFactor) * tvo

	if sim.CompromisedRSUBitMap.Get(int(rsu.Id)) {
		// randomly do a kind of evil
		switch sim.R.RandIntRange(0, dtm.DropPositiveTrustValueOffset) {
		case dtm.FlipTrustValueOffset:
			res = -res
		case dtm.DropPositiveTrustValueOffset:
			if res > 0 {
				res = 0
			}
		}
	}
	return res
}
