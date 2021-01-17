package dtm

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/rpc/pb"
	"github.com/pga2rn/ib-dtm_framework/shared"
	"github.com/pga2rn/ib-dtm_framework/shared/fwtype"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"sync"
)

// factors for generating trust values
// 1. raw trust value offsets(from iterating through RSU)
// 2. time factor function(genesis, slot, trace back epoch)
// 3. compromised RSU? (compromised RSU bitmap, experiments setup)
// 4. epoch length(from simConfig)

// init the storage area
func (session *DTMLogicSession) initDataStructureForEpoch(epoch uint32) {
	logutil.GetLogger(PackageName).Debugf("[initDataStructureForEpoch] epoch %v", epoch)
	for expName := range session.ExpConfig {
		head := session.TrustValueStorageHead[expName]
		if _, err := head.InitTrustValueStorageObject(epoch, session.SimConfig); err != nil {
			logutil.GetLogger(PackageName).
				Fatalf("[initDataStructureForEpoch] failed to allocate storage, expName %v", expName)
		}
	}
}

func (session *DTMLogicSession) calculateTrustValueHelper(
	tvo *fwtype.TrustValueOffset,
	compromisedRSUFlag bool,
) float32 {
	// tuned the trust value offset with weight and time factor
	res := tvo.TrustValueOffset * tvo.Weight /
		float32(session.SimConfig.SlotsPerEpoch)

	// compromised RSU will alter the raw trust value offsets:
	if compromisedRSUFlag {
		switch tvo.AlterType {
		case fwtype.Flipped:
			res = -res
		case fwtype.Dropped:
			res = 0
		}
	}
	return res
}

func (session *DTMLogicSession) genProposalTrustValue(ctx context.Context, epoch uint32) {
	logutil.GetLogger(PackageName).Debugf("[genProposalTrustValue] start to process for epoch %v", epoch)
	// dial to the IB-DTM module
	select {
	case <-ctx.Done():
		logutil.GetLogger(PackageName).Fatalf("[genProposalTrustValue] context canceled")
	default:
		// signal the ib-dtm
		session.ChanIBDTM <- true

		// wait for results from the ib-dtm module
		for {
			v := <-session.ChanIBDTM
			switch v.(type) {
			case shared.IBDTM2DTMCommunication:
				pack := v.(shared.IBDTM2DTMCommunication)
				head := session.TrustValueStorageHead[pack.ExpName]

				// get the head block of the trust value storage chain
				headBlock := head.GetHeadBlock()

				if err := headBlock.SetTrustValueList(pack.Epoch, pack.TrustValueList); err != nil {
					logutil.GetLogger(PackageName).Fatalf("[genProposalTrustValue] failed for exp %v, epoch %v", pack.ExpName, epoch)
				}
			case bool:
				// finish transmitting all experiments
				return
			}
		}
	}
}

func (session *DTMLogicSession) genBaselineTrustValue(ctx context.Context, epoch uint32) {
	select {
	case <-ctx.Done():
		logutil.GetLogger(PackageName).Fatalf("[genBaselineTrustValue] context canceled")
	default:
		// for each RSU
		RSUwg := sync.WaitGroup{}

		// iterate all RSU
		for i := 0; i < session.SimConfig.RSUNum; i++ {
			x, y := session.SimConfig.IndexToCoord(uint32(i))
			r := session.RSUs[x][y]

			RSUwg.Add(1)

			// for each experiment
			expWg := sync.WaitGroup{}
			go func() {
				for expName, exp := range session.ExpConfig {
					if exp.Type == pb.ExperimentType_PROPOSAL {
						continue
					}
					expWg.Add(1)

					// for each pair of trust value offsets, trust value will be calculated for every experiments
					// get the storage head & storage block
					tvStorageHead := session.TrustValueStorageHead[expName]
					tvStorageBlock := tvStorageHead.GetHeadBlock()

					// for each exp, spawn a go routine
					go func(exp *config.ExperimentConfig) {
						baseSlot, currentSlot := uint32(0), (epoch+1)*session.SimConfig.SlotsPerEpoch
						if epoch > uint32(exp.TrustValueOffsetsTraceBackEpochs) {
							baseSlot = currentSlot -
								session.SimConfig.SlotsPerEpoch*uint32(exp.TrustValueOffsetsTraceBackEpochs)
						}

						// RSU: for every slots
						// dive into the slot
						c := make(chan []interface{})
						// define a call back function to take the value out of sync.map
						f := func(key, value interface{}) bool {
							c <- []interface{}{key, value}
							return true
						}

						// the following routine will capture the key and value from the sync map
						go func() {
							for pair := range c {
								key, value := pair[0].(uint32), pair[1].(*fwtype.TrustValueOffset)
								if key != value.VehicleId {
									logutil.GetLogger(PackageName).
										Debugf("[genBaselineTrustValue] mismatch vid! %v in vehicle and %v in tvo", key, value.VehicleId)
									continue // ignore invalid trust value offset record
								}

								// if the trust value offset is forged, and cRSU setting is not activated
								// the tvo will not be counted
								if !exp.CompromisedRSUFlag && value.AlterType == fwtype.Forged {
									continue
								}

								// whether to respect compromisedRSU assignment or not
								compromisedRSUFlag := session.CompromisedRSUBitMap.Get(int(r.Id)) && exp.CompromisedRSUFlag

								// generate!
								res := session.calculateTrustValueHelper(value, compromisedRSUFlag)
								// add value to the storage block
								tvStorageBlock.AddTrustRatingForVehicle(key, res)
							} // receiving data from sync map
						}() // go routine

						// emit paris within the sync map
						for slotIndex := baseSlot; slotIndex < currentSlot; slotIndex++ {
							// get the slot (a sync map)
							slotInstance := r.GetSlotInRing(slotIndex)

							// iterate through the all slots in sync.Map
							slotInstance.Range(f)
						}
						close(c)
						expWg.Done()
					}(exp) // exp go routine
				} // exp for loop
				RSUwg.Done()
			}() // RSU go routine
			expWg.Wait()
		} // RSU for loop
		RSUwg.Wait()
	} // context
}

// iterate through the trust value storage for the specific epoch
// flag out the misbehaving vehicles accordingly
// trust value below 0 will be treated as misbehaving
func (session *DTMLogicSession) flagMisbehavingVehicles(ctx context.Context, epoch uint32) {
	logutil.GetLogger(PackageName).Debugf("[flagMisbehavingVehicles] epoch %v", epoch)
	defer logutil.GetLogger(PackageName).
		Debugf("[flagMisbehavingVehicles] epoch %v Done", epoch)

	select {
	case <-ctx.Done():
		logutil.GetLogger(PackageName).Debugf("[flagMisbehavingVehicles] context canceled")
		return
	default:
		// iterate through every experiment's data storage
		for expName := range session.ExpConfig {

			// get the head of the storage
			head := session.TrustValueStorageHead[expName]
			headBlock := head.GetHeadBlock()
			ep, list, bmap := headBlock.GetTrustValueList()
			if ep != epoch {
				logutil.GetLogger(PackageName).Debugf("[flagMisbehavingVehicles] epoch mismatch! ep %v, epoch %v", ep, epoch)
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
					if vid, tv := pair[0].(uint32), pair[1].(float32); tv < 0 {
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
