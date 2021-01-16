package ib_dtm

import (
	"context"
	"errors"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/shared"
	"github.com/pga2rn/ib-dtm_framework/shared/fwtype"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"sync"
)

func (session *IBDTMSession) processEpoch(ctx context.Context, slot uint32) {
	epoch := uint32(0)
	if slot != 0 {
		epoch = slot/session.IBDTMConfig.SlotsPerEpoch - 1
	}

	logutil.GetLogger(PackageName).Debugf("[processEpoch] epoch %v", epoch)
	defer logutil.GetLogger(PackageName).Debugf("[processEpoch] epoch %v, done", epoch)

	select {
	case <-ctx.Done():
		logutil.GetLogger(PackageName).Fatalf("[processEpoch] context canceled")
	default:
		switch slot {
		case uint32(0):
			for _, bs := range session.BeaconStatus {
				// generate shuffledIdList for next epoch
				bs.UpdateShardStatus(ctx, epoch)
			}
		default:
			// generate trust value for the epoch, for all experiments
			session.genTrustValue(ctx, epoch)
			// inform the dtm module with latest trust value
			session.WaitForDTModule(ctx, epoch)

			// for each experiment, execute ib-dtm logics
			for _, bs := range session.BeaconStatus {
				// update each rsu's balance
				bs.ProcessBalanceAdjustment(ctx, epoch)
				// filter out rsus with insufficient balance
				bs.ProcessLiveCycle(ctx, epoch)
				// generate shuffledIdList for next epoch
				bs.UpdateShardStatus(ctx, epoch+1)
			}
		}
	}
}

func (session *IBDTMSession) ProcessSlot(ctx context.Context, slot uint32) {
	logutil.GetLogger(PackageName).Debugf("[processSlot] slot %v", slot)
	defer logutil.GetLogger(PackageName).Debugf("[processSlot] slot %v, done", slot)

	// wait for sim signal
	<-session.ChanSim

	select {
	case <-ctx.Done():
		logutil.GetLogger(PackageName).Fatalf("[processSlot] context canceled")
	default:
		// for each experiments, each exp has a beaconstatus and a chain
		for _, exp := range session.ExpConfigList {
			bs := session.BeaconStatus[exp.Name]
			beaconBlock := session.Blockchain[exp.Name].GetHeadBlock()

			// for each shard block
			for shardId := 0; shardId < session.IBDTMConfig.ShardNum; shardId++ {
				// committee is one-to-one mapping to the slot index
				shardBlock := beaconBlock.shards[shardId]
				cid := slot % bs.IBDTMConfig.SlotsPerEpoch
				proposerId := bs.shardStatus[shardId].proposer[cid]
				proposerValidator := bs.validators.Validators[proposerId]
				shardBlock.proposer = proposerId

				// first we let the proposer to propose the block
				// check if the proposer is active
				if !bs.IsValidatorActive(proposerId) {
					continue // block skipped
				}

				// mapping the validator to the RSU
				x, y := session.SimConfig.IndexToCoord(proposerId)
				proposerRSU := session.RSUs[x][y]
				//logutil.GetLogger(PackageName).Infof("[processSlot] slot %v, shard %v, proposer %v", slot, shardId, proposerId)

				// get the trust value offsets list
				startSlot, endSlot := proposerValidator.GetNextSlotForUpload(), slot
				for i := startSlot; i <= endSlot; i++ {
					tvolist := proposerRSU.GetSlotInRing(i)
					if tvolist == nil {
						logutil.GetLogger(PackageName).Debugf("[processSlot] nil tvolist, slot %v, rsu %v", slot, proposerId)
					}

					// try to save a new copy of tvolist
					shardBlock.tvoList[i] = tvolist
					//logutil.GetLogger(PackageName).Infof("[processSlot] s%v, sd %v, pr %v, tvo %v", slot, shardId, proposerId, tmp)
				}
				// update the next available update slot
				proposerValidator.SetNextSlotForUpload(endSlot + 1)

				// for each member in the committee,
				// start to vote for the new block
				committee := bs.GetCommitteeByCommitteeId(uint32(shardId), cid)

				switch exp.CompromisedRSUFlag {
				case true:
					proposerIsCompromised := session.CompromisedRSUBitMap.Get(int(proposerId))
					for index, vid := range committee {
						if !bs.IsValidatorActive(vid) {
							continue
						}
						validatorIsCompromised := session.CompromisedRSUBitMap.Get(int(vid))

						// TODO: implement real validation here
						rn := bs.R.Float32()

						switch {
						// the bad voter will let the bad RSU propose
						case proposerIsCompromised && validatorIsCompromised:
							switch {
							case rn < 0.7:
								shardBlock.votes[index] = true
							default:
								shardBlock.votes[index] = false
							}
						// the good voter will not let the bad RSU go
						case proposerIsCompromised && !validatorIsCompromised:
							switch {
							case rn < 0.9:
								shardBlock.votes[index] = false
							default:
								shardBlock.votes[index] = true
							}
						// good validator will vote for good RSU
						case !proposerIsCompromised && !validatorIsCompromised:
							shardBlock.votes[index] = true
						// bad validator will camouflage itself by voting for good RSU
						case !proposerIsCompromised && validatorIsCompromised:
							switch {
							case rn < 0.6:
								shardBlock.votes[index] = true
							default:
								shardBlock.votes[index] = false
							}

						}
					} // voting
				case false:
					// everyone is honest
					for index, _ := range committee {
						shardBlock.votes[index] = true
					} // voting
				}

				// PGA2RN: end of rewrite

				// check the voting stakes to decide whether the block should pass or not
				totalStake, gainedStake := float32(0), float32(0)
				for index, approved := range shardBlock.votes {
					vid := committee[index]
					if !bs.IsValidatorActive(uint32(index)) {
						continue
					}

					totalStake += bs.validators.Validators[int(vid)].effectiveStake
					if approved {
						gainedStake += bs.validators.Validators[int(vid)].effectiveStake
					}
				}
				// if the block gained enough stakes
				if gainedStake > totalStake*2/3 {
					shardBlock.skipped = false
				}

			} // shard block
		} // each experiments
		session.ChanSim <- true
	}
}

func (session *IBDTMSession) WaitForDTModule(ctx context.Context, epoch uint32) {
	logutil.GetLogger(PackageName).Debugf("[WaitForDTModule] epoch %v", epoch)
	defer logutil.GetLogger(PackageName).Debugf("[WaitForDTModule] epoch %v, done", epoch)

	select {
	case <-ctx.Done():
		logutil.GetLogger(PackageName).Fatalf("[WaitForDTModule] context canceled")
	default:
		<-session.ChanDTM // wait for signal to transmit the data

		for expName, data := range session.TrustValueStorage {
			res := shared.IBDTM2DTMCommunication{
				Epoch:          epoch,
				ExpName:        expName,
				TrustValueList: data,
			}
			session.ChanDTM <- res
		}
		session.ChanDTM <- true // finish transmission
	}
}

func (session *IBDTMSession) WaitForSimulator(ctx context.Context) error {
	defer logutil.GetLogger(PackageName).Debugf("[WaitForSimulator] init finished!")
	select {
	case <-ctx.Done():
		return errors.New("[WaitForSimulator] context canceled")
	case v := <-session.ChanSim:
		// unpack
		pack := v.(shared.SimInitIBDTMCommunication)

		session.RSUs = pack.RSUs
		session.CompromisedRSUBitMap = pack.CompromisedRSUBitMap
		session.rmu = pack.Rmu

		// after the init, signal the simulator
		session.ChanSim <- true
		return nil
	}

}

func (bs *BeaconStatus) ProcessBalanceAdjustment(ctx context.Context, epoch uint32) {
	logutil.GetLogger(PackageName).Debugf("[ProcessBalanceAdjustment] epoch %v", epoch)
	defer logutil.GetLogger(PackageName).Debugf("[ProcessBalanceAdjustment] epoch %v, done", epoch)

	select {
	case <-ctx.Done():
		logutil.GetLogger(PackageName).Fatalf("[ProcessBalanceAdjustment] context canceled")
	default:
		// iterate through the beaconblocks for each slots in the epoch
		start, end := epoch*bs.IBDTMConfig.SlotsPerEpoch, (epoch+1)*bs.IBDTMConfig.SlotsPerEpoch
		for i := start; i < end; i++ { // for each slots
			beaconblock := bs.Blockchain.GetBlockForSlot(i)
			for shardId, block := range beaconblock.shards { // for each shard
				proposer := bs.validators.Validators[block.proposer]

				if block.skipped {
					logutil.GetLogger(PackageName).Debugf("block skipped at slot %v, shard %v", block.slot, shardId)
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
	logutil.GetLogger(PackageName).Debugf("[ProcessLiveCycle] epoch %v", epoch)
	defer logutil.GetLogger(PackageName).Debugf("[ProcessLiveCycle] epoch %v, done", epoch)

	select {
	case <-ctx.Done():
		logutil.GetLogger(PackageName).Fatalf("[ProcessLiveCycle] context canceled")
	default:
		// check stake
		for _, validator := range bs.validators.Validators {
			if bs.IsValidatorActive(validator.Id) && validator.effectiveStake < bs.IBDTMConfig.EffectiveStakeLowerBound {
				//logutil.GetLogger(PackageName).Warnf("[lifecycle] r %v has been inactivated", validator.Id)
				bs.InactivateValidator(validator.Id)
			}
			// debug, show all validator's stake
			//logutil.GetLogger(PackageName).Infof("[lifecycle]r %v, es %v, its %v", validator.Id, validator.effectiveStake, validator.itsStake.GetAmount())
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
	logutil.GetLogger(PackageName).Debugf("[genTrustValue] epoch %v", epoch)

	// iterate through the blockchain for all experiments
	upperWg := sync.WaitGroup{}
	for _, exp := range session.ExpConfigList {
		// init storage area
		session.TrustValueStorage[exp.Name] = &fwtype.TrustValuesPerEpoch{}
		blockchain := session.Blockchain[exp.Name]
		session.TrustValueStorage[exp.Name] = &fwtype.TrustValuesPerEpoch{}
		result := session.TrustValueStorage[exp.Name]

		upperWg.Add(1)
		go func(exp *config.ExperimentConfig) {
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

				for _, shard := range block.shards { // spawn go routines for each shard in each slots
					wg.Add(1)
					go func(shard *ShardBlock) {
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
								//logutil.GetLogger(PackageName).Infof("[genTrustValue] e %v, sd %v, k %v, v %v", epoch, shardId, key, value.TrustValueOffset)

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
						wg.Done() // job done for shards
					}(shard)
				} // iterate shards
			} // iterate slots for loop

			wg.Wait()
			upperWg.Done()
		}(exp) // go routine for each experiment

		upperWg.Wait()
	} // experiment
}
