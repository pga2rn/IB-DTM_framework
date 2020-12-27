package dtm

//
//func (sim *SimulationSession) genTrustValue(ctx context.Context, slot uint64) {
//	logutil.LoggerList["core"].Debugf("[genTrustValue] start to process for epoch %v", slot/sim.Config.SlotsPerEpoch)
//	defer logutil.LoggerList["core"].
//		Debugf("[genTrustValue] epoch %v, slot %v done", slot/sim.Config.SlotsPerEpoch, slot)
//
//	select {
//	case <-ctx.Done():
//		return
//	default:
//		if slot != timeutil.SlotsSinceGenesis(sim.Config.Genesis) {
//			logutil.LoggerList["core"].
//				Warnf("[genTrustValue] mismatch slot index! potential async")
//		}
//		if slot%sim.Config.SlotsPerEpoch != 0 {
//			logutil.LoggerList["core"].Fatalf("[genTrustValue] call func at non checkpoint slot, abort")
//		}
//
//		// init a data structure to store the trust value
//		trustValueRecord :=
//			dtmtype.InitTrustValueStorageObject(slot / sim.Config.SlotsPerEpoch)
//
//		wg := sync.WaitGroup{}
//
//		// iterate all RSU
//		// set deadline
//		epochCtx, cancel := context.WithDeadline(ctx, timeutil.NextEpochTime(sim.Config.Genesis, slot))
//		for x := range sim.RSUs {
//			for y := range sim.RSUs[x] {
//				rsu := sim.RSUs[x][y]
//
//				// use go routines to collect every RSU's data
//				// add one worker to wait group
//				wg.Add(1)
//				go func() {
//					select {
//					case <-epochCtx.Done():
//						logutil.LoggerList["core"].Fatalf("[genTrustValue] times up for collecting RSU data at the end of epoch, abort")
//						return
//					default:
//						// RSU: for every slots
//						for slotIndex := 0; slotIndex < int(sim.Config.SlotsPerEpoch); slotIndex++ {
//							// get the slot
//							slotInstance := rsu.TrustValueOffsetPerSlot[slotIndex]
//							// dive into the slot
//							for vid, tvo := range slotInstance {
//								if vid != tvo.VehicleId {
//									logutil.LoggerList["core"].
//										Warnf("[genTrustValue] mismatch vid! %v in vehicle and %v in tvo", vid, tvo.VehicleId)
//									continue // ignore invalid trust value offset record
//								}
//
//								//
//								tunedTrustValueOffset := sim.genTrustValueHelper(rsu, tvo.TrustValueOffset, slot)
//								if op, ok := trustValueRecord.TrustValueList.LoadOrStore(vid, tunedTrustValueOffset); ok {
//									// ok means there is already value stored in place
//									// the existed value is loaded to variable op
//									trustValueRecord.TrustValueList.Store(vid, op.(float32)+tunedTrustValueOffset)
//								}
//							}
//						}
//					} // select
//					wg.Done() // job done,
//				}() // go routine
//			}
//		}
//
//		// wait for all work to finish their job
//		wg.Wait()
//		// after all the workers finish their job, cancel the context
//		cancel()
//
//		// TODO: realize background tracking services to keep records of trust value
//		sim.TrustValueList = trustValueRecord.TrustValueList
//	}
//}
//
//
//// return tuned trust value offset!(may be or may not be compromised!)
//func (sim *SimulationSession) genTrustValueHelper(rsu *dtmLogic.RSU, tvo float32, slot uint64) float32 {
//	timeFactor := timefactor.GetTimeFactor(
//		sim.Config.TimeFactorType,
//		sim.Config.Genesis,
//		timeutil.SlotStartTime(sim.Config.Genesis, slot),
//		timeutil.NextEpochTime(sim.Config.Genesis, slot),
//	)
//	res := float32(timeFactor) * tvo
//
//	if sim.CompromisedRSUBitMap.Get(int(rsu.Id)) {
//		// randomly do a kind of evil
//		switch sim.R.RandIntRange(0, dtmLogic.DropPositiveTrustValueOffset) {
//		case dtmLogic.FlipTrustValueOffset:
//			res = -res
//		case dtmLogic.DropPositiveTrustValueOffset:
//			if res > 0 {
//				res = 0
//			}
//		}
//	}
//	return res
//}
//
