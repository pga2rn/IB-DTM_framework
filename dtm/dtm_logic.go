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

// init the storage area
func (session *DTMLogicSession) initDataStructureForEpoch(epoch uint64) {
	logutil.LoggerList["dtm"].Debugf("[initDataStructureForEpoch] epoch %v", epoch)
	for expName := range *session.Config {
		head := (*session.TrustValueStorageHead)[expName]
		if _, err := head.InitTrustValueStorageObject(epoch, session.SimConfig); err != nil {
			logutil.LoggerList["dtm"].
				Fatalf("[initDataStructureForEpoch] failed to allocate storage, expName %v", expName)
		}
	}
}

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

	// compromised RSU will do the following evils:
	// 1. flip the trust value
	// 2. drop trust value offset: 0.3
	if compromisedRSUFlag {
		switch tvo.AlterType {
		case dtmtype.Flipped:
			res = -res
		case dtmtype.Dropped:
			res = 0
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

func (session *DTMLogicSession) genTrustValue(ctx context.Context, epoch uint64) {
	logutil.LoggerList["dtm"].Debugf("[genTrustValue] start to process for epoch %v", epoch)
	defer logutil.LoggerList["dtm"].
		Debugf("[genTrustValue] epoch %v done", epoch)

	select {
	case <-ctx.Done():
		logutil.LoggerList["dtm"].Fatalf("[genTrustValue] context canceled")
		return
	default:
		// for go routine
		wg := sync.WaitGroup{}

		// iterate all RSU
		for x := range *session.RSUs {
			for y := range (*session.RSUs)[x] {
				session.rmu.Lock()
				r := (*session.RSUs)[x][y]
				session.rmu.Unlock()

				// use go routines to collect every RSU's data
				// add one worker to wait group
				wg.Add(1)
				go func() {
					select {
					case <-ctx.Done():
						logutil.LoggerList["simulator"].Fatalf("[genTrustValue] times up for collecting RSU data at the end of epoch, abort")
						return
					default:
						baseSlot, currentSlot := r.GetRingInformation()

						// RSU: for every slots
						for slotIndex := baseSlot; slotIndex <= currentSlot; slotIndex++ {
							// get the slot (a sync map)
							slotInstance := r.GetSlotInRing(slotIndex)

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
								case <-ctx.Done():
									logutil.LoggerList["simulator"].Fatalf("[genTrustValue] times up for collecting RSU data at the end of epoch, abort")
									return
								default:
									for pair := range c {
										key, value := pair[0].(uint64), pair[1].(*dtmtype.TrustValueOffset)
										if key != value.VehicleId {
											logutil.LoggerList["simulator"].
												Warnf("[genTrustValue] mismatch vid! %v in vehicle and %v in tvo", key, value.VehicleId)
											continue // ignore invalid trust value offset record
										}

										// for each pair of trust value offsets, trust value will be calculated for every experiments
										for expName, exp := range *session.Config {
											// get the storage head & storage block
											tvStorageHead := (*session.TrustValueStorageHead)[expName]
											tvStorageBlock := tvStorageHead.GetHeadBlock()

											// whether to respect compromisedRSU assignment or not
											compromisedRSUFlag := session.CompromisedRSUBitMap.Get(int(r.Id)) && exp.CompromisedRSUFlag

											// generate!
											tfactor := session.genTimeFactorHelper(expName, slotIndex)
											res := session.genTrustValueHelper(
												float32(tfactor), value,
												compromisedRSUFlag, exp.TimeFactorFlag,
											)
											// add value to the storage block
											tvStorageBlock.AddValue(key, res)
										} // experiment loop
									} // receiving data from sync map
								} // context
							}() // go routine

							// iterate through the all slots in sync.Map
							slotInstance.Range(f)
							close(c)
						}
					} // select
					wg.Done() // job done,
				}() // go routine
			} // iterate RSUs inner for loop
		} // outer for loop

		// wait for all work to finish their job
		wg.Wait()
	}
}

// iterate through the trust value storage for the specific epoch
// flag out the misbehaving vehicles accordingly
// trust value below 0 will be treated as misbehaving
func (session *DTMLogicSession) flagMisbehavingVehicles(ctx context.Context, epoch uint64) {
	logutil.LoggerList["dtm"].Debugf("[flagMisbehavingVehicles] epoch %v", epoch)
	defer logutil.LoggerList["dtm"].
		Debugf("[flagMisbehavingVehicles] epoch %v done", epoch)

	select {
	case <-ctx.Done():
		logutil.LoggerList["dtm"].Debugf("[flagMisbehavingVehicles] context canceled")
		return
	default:
		// iterate through every experiment's data storage
		for expName := range *session.Config {
			// get the head of the storage
			head := (*session.TrustValueStorageHead)[expName]
			headBlock := head.GetHeadBlock()
			ep, list, bmap := headBlock.GetTrustValueList()
			if ep != epoch {
				logutil.LoggerList["dtm"].Debugf("[flagMisbehavingVehicles] epoch mismatch! ep %v, epoch %v", ep, epoch)
				return
			}

			// iterate through sync.Map
			c := make(chan []interface{})
			f := func(key, value interface{}) bool {
				c <- []interface{}{key, value}
				return true
			}

			// flag misbehaving vehicles
			go func() {
				for pair := range c {
					if vid, tv := pair[0].(uint64), pair[1].(float32); tv < 0 {
						bmap.Set(int(vid), true)
					}
				}
			}()

			// iterate via Range
			list.Range(f)
			close(c)
		}
	}
}
