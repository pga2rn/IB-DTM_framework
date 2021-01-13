package ib_dtm

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/fwtype"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"sync"
)

func (bs *BeaconStatus) ProcessBalanceAdjustment(ctx context.Context, epoch uint32) {
	logutil.LoggerList["ib-dtm"].Debugf("[ProcessBalanceAdjustment] epoch %v", epoch)
	defer logutil.LoggerList["ib-dtm"].Debugf("[ProcessBalanceAdjustment] epoch %v, done", epoch)

	select {
	case <-ctx.Done():
		logutil.LoggerList["ib-dtm"].Fatalf("[ProcessBalanceAdjustment] context canceled")
	default:
		// iterate through the beaconblocks for each slots in the epoch
		start, end := epoch*bs.IBDTMConfig.SlotsPerEpoch, (epoch+1)*bs.IBDTMConfig.SlotsPerEpoch
		for i := start; i < end; i++ { // for each slots
			beaconblock := bs.Blockchain.GetBlockForSlot(i)
			for shardId, block := range beaconblock.shards { // for each shard
				proposer := bs.validators.Validators[block.proposer]

				if block.skipped {
					logutil.LoggerList["ib-dtm"].Debugf("block skipped at slot %v, shard %v", block.slot, shardId)
					proposer.AddEffectiveStake(-1 * bs.IBDTMConfig.BaseReward * bs.IBDTMConfig.PenaltyFactor)
				} else {
					factor := bs.GetRewardFactor(proposer.Id)
					proposer.AddEffectiveStake(bs.IBDTMConfig.BaseReward * factor)
					// add its stake
					count := 0
					f := func(key, value interface{}) bool {
						count++
						return true
					}
					for _, tvolist := range block.tvoList {
						tvolist.Range(f)
					}
					proposer.AddITStake(epoch, float32(count))
				}

				// scan through the committee, give reward and penalty to each validators accordingly
				// get the slot committee(committee is one-to-one map to the slots)
				committee := bs.GetCommitteeByCommitteeId(uint32(shardId), block.slot%bs.SimConfig.SlotsPerEpoch)
				// iterate throught the committee
				for index, vid := range committee {
					// not counting the proposer and inactive validator
					if vid == block.proposer || !bs.IsValidatorActive(vid) {
						continue
					}
					factor := bs.GetRewardFactor(vid)
					// check the vote and the block approval
					switch {
					// if the block is not valid, but the voter votes for it
					case block.skipped == true && block.votes[index] == true:
						bs.validators.Validators[vid].AddEffectiveStake(-1 * bs.IBDTMConfig.BaseReward)
					// if the block is valid, and the voter votes for it
					case block.skipped == false && block.votes[index] == true:
						bs.validators.Validators[vid].AddEffectiveStake(bs.IBDTMConfig.BaseReward * factor)
					}
				}

			}
		}
	}
}

func (bs *BeaconStatus) ProcessLiveCycle(ctx context.Context, epoch uint32) {
	logutil.LoggerList["ib-dtm"].Debugf("[ProcessLiveCycle] epoch %v", epoch)
	defer logutil.LoggerList["ib-dtm"].Debugf("[ProcessLiveCycle] epoch %v, done", epoch)

	select {
	case <-ctx.Done():
		logutil.LoggerList["ib-dtm"].Fatalf("[ProcessLiveCycle] context canceled")
	default:
		// check stake
		for _, validator := range bs.validators.Validators {
			if bs.IsValidatorActive(validator.Id) && validator.effectiveStake < bs.IBDTMConfig.EffectiveStakeLowerBound {
				//logutil.LoggerList["ib-dtm"].Warnf("[lifecycle] r %v has been inactivated", validator.Id)
				bs.InactivateValidator(validator.Id)
			}
			// debug, show all validator's stake
			//logutil.LoggerList["ib-dtm"].Infof("[lifecycle]r %v, es %v, its %v", validator.Id, validator.effectiveStake, validator.itsStake.GetAmount())
		}

		// check slash
		for slashedValidator, _ := range bs.slashings {
			bs.InactivateValidator(slashedValidator)
		}
	}
}

func (session *IBDTMSession) calculateTrustValueHelper(
	tvo *fwtype.TrustValueOffset,
	compromisedRSUFlag bool) float32 {

	res := tvo.Weight * tvo.TrustValueOffset / float32(session.SimConfig.SlotsPerEpoch)

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

func (session *IBDTMSession) genTrustValue(ctx context.Context, epoch uint32) {
	logutil.LoggerList["ib-dtm"].Debugf("[genTrustValue] epoch %v", epoch)

	// iterate through the blockchain for all experiments
	for _, exp := range session.ExpConfigList {
		// init storage area
		session.TrustValueStorage[exp.Name] = &fwtype.TrustValuesPerEpoch{}
		blockchain := session.Blockchain[exp.Name]
		session.TrustValueStorage[exp.Name] = &fwtype.TrustValuesPerEpoch{}
		result := session.TrustValueStorage[exp.Name]

		startSlot, endSlot := epoch*session.SimConfig.SlotsPerEpoch, (epoch+1)*session.SimConfig.SlotsPerEpoch
		if epoch < uint32(exp.TrustValueOffsetsTraceBackEpochs) {
			startSlot = 0
		} else {
			startSlot = endSlot - uint32(exp.TrustValueOffsetsTraceBackEpochs)*session.SimConfig.SlotsPerEpoch
		}

		// iterate through each slots
		wg := sync.WaitGroup{}
		// for each shard
		for i := startSlot; i < endSlot; i++ {
			block := blockchain.GetBlockForSlot(i)

			wg.Add(1)
			go func(block *BeaconBlock, result *fwtype.TrustValuesPerEpoch) {
				// spawn go routines for each slots
				// for each shard
				for _, shard := range block.shards {
					// dive into the slot
					c := make(chan []interface{})
					// define a call back function to take the value out of sync.map
					f := func(key, value interface{}) bool {
						c <- []interface{}{key, value}
						return true
					}

					// capture all values in the slot
					go func() {
						for pair := range c {
							key, value := pair[0].(uint32), pair[1].(*fwtype.TrustValueOffset)
							//logutil.LoggerList["ib-dtm"].Infof("[genTrustValue] e %v, sd %v, k %v, v %v", epoch, shardId, key, value.TrustValueOffset)

							if key != value.VehicleId {
								logutil.LoggerList["simulator"].
									Debugf("[genBaselineTrustValue] mismatch vid! %v in vehicle and %v in tvo", key, value.VehicleId)
								continue // ignore invalid trust value offset record
							}

							// if the trust value offset is forged, and cRSU setting is not activated
							// the tvo will not be counted
							if !exp.CompromisedRSUFlag && value.AlterType == fwtype.Forged {
								continue
							}

							// whether the proposer is compromised RSU & enable compromised flag
							compromisedRSUFlag := session.CompromisedRSUBitMap.Get(int(shard.proposer)) && exp.CompromisedRSUFlag

							tvo := session.calculateTrustValueHelper(value, compromisedRSUFlag)
							if op, ok := result.LoadOrStore(value.VehicleId, tvo); ok {
								result.Store(value.VehicleId, tvo+op.(float32))
							}
						}
					}()

					// a block main contain many slots' trust value offsets
					for _, tvoList := range shard.tvoList {
						tvoList.Range(f)
					}
					close(c)
				} // iterate shards
				wg.Done() // job done for shards
			}(block, result) // go routine for each slot
		} // iterate slots for loop

		wg.Wait()
	} // experiment
}
