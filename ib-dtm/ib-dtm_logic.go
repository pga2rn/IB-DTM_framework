package ib_dtm

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/fwtype"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"sync"
)

func (bs *BeaconStatus) ProcessBalanceAdjustment(ctx context.Context, epoch uint32) {
	if bs.Epoch != epoch && epoch != 0 {
		logutil.LoggerList["ib-dtm"].Fatalf("[ProcessForNewEpoch] epoch async %v, %v", bs.Epoch, epoch)
	}

	select {
	case <-ctx.Done():
		logutil.LoggerList["ib-dtm"].Fatalf("[ProcessBalanceAdjustment] context canceled")
	default:
		// iterate through the beaconblocks for each slots in the epoch
		start, end := epoch*bs.IBDTMConfig.SlotsPerEpoch, (epoch+1)*bs.IBDTMConfig.SlotsPerEpoch
		for i := start; i < end; i++ {
			beaconblock := bs.Blockchain.GetBlockForSlot(i)
			for shardId, block := range beaconblock.shards {
				proposer := bs.validators[block.proposer]

				if block.skipped {
					proposer.AddITStake(-1 * bs.IBDTMConfig.BaseReward * bs.IBDTMConfig.PenaltyFactor)
				} else {
					factor := bs.GetRewardFactor(proposer.Id)
					proposer.AddITStake(bs.IBDTMConfig.BaseReward * factor)
				}
				// scan through the committee, give reward and penalty to each validators accordingly
				// get the slot committee
				committee := bs.GetCommitteeByValidatorId(uint32(shardId), proposer.Id)
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
						bs.validators[vid].AddITStake(-1 * bs.IBDTMConfig.BaseReward)
					// if the block is valid, and the voter votes for it
					case block.skipped == false && block.votes[index] == true:
						bs.validators[vid].AddITStake(bs.IBDTMConfig.BaseReward * factor)
					}
				}

				// TODO: move whistleblowing here
				// TODO: move slashing logics here
			}
		}
	}
}

func (bs *BeaconStatus) ProcessLiveCycle(ctx context.Context, epoch uint32) {
	if bs.Epoch != epoch && epoch != 0 {
		logutil.LoggerList["ib-dtm"].Fatalf("[ProcessLiveCycle] epoch async %v, %v", bs.Epoch, epoch)
	}

	select {
	case <-ctx.Done():
		logutil.LoggerList["ib-dtm"].Fatalf("[ProcessLiveCycle] context canceled")
	default:
		// check stake
		for _, validator := range bs.validators {
			if validator.effectiveStake < bs.IBDTMConfig.EffectiveStakeLowerBound {
				bs.InactivateValidator(validator.Id)
			}
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
	// iterate through the blockchain for all experiments
	for _, exp := range session.ExpConfigList {
		// init storage area
		session.TrustValueStorage[exp.Name] = &fwtype.TrustValuesPerEpoch{}
		blockchain := session.Blockchain[exp.Name]
		result := session.TrustValueStorage[exp.Name]

		startSlot, endSlot := uint32(0), (epoch+1)*session.SimConfig.SlotsPerEpoch
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
			go func() { // spawn go routines for each slots
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
							if key != value.VehicleId {
								logutil.LoggerList["simulator"].
									Warnf("[genBaselineTrustValue] mismatch vid! %v in vehicle and %v in tvo", key, value.VehicleId)
								continue // ignore invalid trust value offset record
							}

							tvo := session.calculateTrustValueHelper(value, exp.CompromisedRSUFlag)
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
			}() // iterate slots
			wg.Done()
		}

		wg.Wait()
	}
}
