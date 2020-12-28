package dtm

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/dtmtype"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/timefactor"
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
	"sync"
	"time"
)

// factors for generating trust values
// 1. raw trust value offsets(from iterating through RSU)
// 2. time factor function(genesis, slot, trace back epoch)
// 3. compromised RSU? (compromised RSU bitmap, experiments setup)
// 4. epoch length(from simConfig)

func (session *DTMLogicSession) genTrustValueHelper(
	tfactor float32,
	tvo *dtmtype.TrustValueOffset,
	compromisedRSUFlag bool,
	timeFactorFlag bool,
) float32 {

	// time factor
	if !timeFactorFlag {
		tfactor = float32(1)
	}

	// tuned the trust value offset with weight and time factor
	res := tvo.TrustValueOffset * tfactor * tvo.Weight /
		float32(session.SimConfig.SlotsPerEpoch)

	// compromised RSU
	if compromisedRSUFlag {
		switch session.R.RandIntRange(0, len(RSUEvilsType)) {
		case FlipTrustValueOffset:
			res = -res
		case DropPositiveTrustValueOffset:
			if res > 0 {
				res = 0
			}
		}
	}

	return res
}

// generate time factor for different experiment setup
func (session *DTMLogicSession) genTimeFactorHelper(name string, slot uint64) float64 {
	var start, end time.Time
	cfg, genesis := (*session.Config)[name], session.SimConfig.Genesis

	slotTime := timeutil.SlotStartTime(genesis, slot)
	epoch := slot / session.SimConfig.SlotsPerEpoch

	if epoch < uint64(cfg.TrustValueOffsetsTraceBackEpochs) {
		// not enough previous epochs for trace back
		start = session.SimConfig.Genesis
		end = timeutil.NextEpochTime(session.SimConfig.Genesis, slot)
	} else {
		start = timeutil.NextEpochTime(
			session.SimConfig.Genesis, epoch-uint64(cfg.TrustValueOffsetsTraceBackEpochs))
		end = timeutil.NextEpochTime(session.SimConfig.Genesis, slot)
	}
	return timefactor.GetTimeFactor(cfg.TimeFactorType, start, slotTime, end)
}

func (session *DTMLogicSession) genTrustValue(ctx context.Context, slot uint64) {
	logutil.LoggerList["dtm"].Debugf("[genTrustValue] start to process for epoch %v", slot/session.SimConfig.SlotsPerEpoch)
	defer logutil.LoggerList["core"].
		Debugf("[genTrustValue] epoch %v, slot %v done", slot/session.SimConfig.SlotsPerEpoch)

	select {
	case <-ctx.Done():
		return
	default:
		if slot != timeutil.SlotsSinceGenesis(session.SimConfig.Genesis) {
			logutil.LoggerList["core"].
				Warnf("[genTrustValue] mismatch slot index! potential async")
		}
		if slot%session.SimConfig.SlotsPerEpoch != 0 {
			logutil.LoggerList["core"].Fatalf("[genTrustValue] call func at non checkpoint slot, abort")
		}

		// for go routine
		wg := sync.WaitGroup{}

		// iterate all RSU
		// set deadline
		epochCtx, cancel := context.WithDeadline(ctx, timeutil.NextEpochTime(session.SimConfig.Genesis, slot))
		for x := range *session.RSUs {
			for y := range (*session.RSUs)[x] {
				session.rmu.Lock()
				rsu := (*session.RSUs)[x][y]
				session.rmu.Unlock()

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
						for slotIndex := 0; slotIndex < int(session.SimConfig.SlotsPerEpoch); slotIndex++ {
							// get the slot (a sync map)
							slotInstance := rsu.GetSlotsInRing(slot)

							// dive into the slot
							c := make(chan []interface{})
							// define a call back function to take the value out of sync.map
							f := func(key, value interface{}) bool {
								c <- []interface{}{key, value}
								return true
							}
							// the following routine will capture the key and value from the sync map
							go func() {
								select {
								case <-epochCtx.Done():
									logutil.LoggerList["core"].Fatalf("[genTrustValue] times up for collecting RSU data at the end of epoch, abort")
									return
								default:
									for pair := range c {
										key, value := pair[0].(uint64), pair[1].(*dtmtype.TrustValueOffset)
										if key != value.VehicleId {
											logutil.LoggerList["core"].
												Warnf("[genTrustValue] mismatch vid! %v in vehicle and %v in tvo", key, value.VehicleId)
											continue // ignore invalid trust value offset record
										}

										// for each pair of trust value offsets, trust value will be calculated for every experiments
										for expName, exp := range *session.Config {
											// get the storage head & storage block
											tvStorageHead := (*session.TrustValueStorageHead)[expName]
											// TODO: we should init the storage block for the epoch before the tv generation function being called
											tvStorageBlock := tvStorageHead.GetHeadBlock()

											// generate!
											tfactor := session.genTimeFactorHelper(expName, slot)
											res := session.genTrustValueHelper(
												float32(tfactor), value,
												exp.CompromisedRSUFlag, exp.TimeFactorFlag,
											)
											// add value to the storage block
											tvStorageBlock.AddValue(key, res)
										} // experiment loop
									} // receiving data from sync map
								} // context
							}() // go routine

							// iterate through the sync.Map
							slotInstance.Range(f)
							close(c)
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
	}
}
