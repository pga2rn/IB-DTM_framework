package ib_dtm

import (
	"context"
	"errors"
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/rpc/pb"
	"github.com/pga2rn/ib-dtm_framework/rsu"
	"github.com/pga2rn/ib-dtm_framework/shared"
	"github.com/pga2rn/ib-dtm_framework/shared/fwtype"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/randutil"
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
	"sync"
)

type IBDTMSession struct {
	IBDTMConfig   *config.IBDTMConfig
	SimConfig     *config.SimConfig
	ExpConfigList []*config.ExperimentConfig

	// chain for every proposal experiments
	Blockchain map[string]*BlockchainRoot

	// pointer to the vehicles and RSU
	// I don't know if it is a good idea to use mutex via pointer
	RSUs                 *[][]*rsu.RSU
	CompromisedRSUBitMap *bitmap.Threadsafe
	rmu                  *sync.Mutex

	// inter-modules communication
	ChanSim, ChanDTM chan interface{}

	Ticker      timeutil.Ticker
	Epoch, Slot uint32

	// beacon status
	BeaconStatus map[string]*BeaconStatus

	// latest trust value results
	TrustValueStorage map[string]*fwtype.TrustValuesPerEpoch

	// a random source
	R *randutil.RandUtil
}

func PrepareBlockchainModule(
	simCfg *config.SimConfig, expCfgList []*config.ExperimentConfig, ibdtmCfg *config.IBDTMConfig,
	sim, dtm chan interface{}) *IBDTMSession {

	session := &IBDTMSession{
		SimConfig:     simCfg,
		ExpConfigList: expCfgList,
		IBDTMConfig:   ibdtmCfg,
		ChanDTM:       dtm,
		ChanSim:       sim,
	}

	session.Blockchain = make(map[string]*BlockchainRoot)
	session.BeaconStatus = make(map[string]*BeaconStatus)

	// init storage area for each experiment
	for _, exp := range expCfgList {
		if exp.Type == pb.ExperimentType_PROPOSAL {
			// init blockchain
			session.Blockchain[exp.Name] = InitBlockchain()

			// init beaconstatus for each experiment
			session.BeaconStatus[exp.Name] = InitBeaconStatus(
				simCfg, ibdtmCfg, session.Blockchain[exp.Name])
		}
	}

	// prepare the ticker
	session.Ticker = timeutil.GetSlotTicker(simCfg.Genesis, simCfg.SecondsPerSlot)

	// prepare the random source
	session.R = randutil.InitRand(123)

	return session
}

func (session *IBDTMSession) processEpoch(ctx context.Context, epoch uint32) {
	select {
	case <-ctx.Done():
		logutil.LoggerList["ib-dtm"].Fatalf("[processEpoch] context canceled")
	default:
		// generate trust value for the previous epoch, for all experiments
		session.genTrustValue(ctx, epoch)
		// inform the dtm module with latest trust value
		go session.dialDTModule(ctx, epoch)

		// for each experiment, execute ib-dtm logics
		for _, bs := range session.BeaconStatus {
			// update each rsu's balance
			bs.ProcessBalanceAdjustment(ctx, epoch)
			// filter out rsus with insufficient balance
			bs.ProcessLiveCycle(ctx, epoch)
			// generate shuffledIdList for next epoch
			bs.UpdateShardStatus(ctx, epoch)
		}
	}
}

func (session *IBDTMSession) ProcessSlot(ctx context.Context, slot uint32) {
	select {
	case <-ctx.Done():
		logutil.LoggerList["ib-dtm"].Fatalf("[processSlo] context canceled")
	default:
		// for each experiments, each exp has a beaconstatus and a chain
		for _, exp := range session.ExpConfigList {
			bs := session.BeaconStatus[exp.Name]
			beaconBlock := session.Blockchain[exp.Name].GetHeadBlock()

			// for each shard block
			for shardId := 0; shardId < session.IBDTMConfig.ShardNum; shardId++ {
				shardBlock := &ShardBlock{
					skipped: true,
					slot:    slot,
				}
				beaconBlock.shards[shardId] = shardBlock
				// committee is one-to-one mapping to the slot index
				cid := slot % bs.IBDTMConfig.SlotsPerEpoch

				// get the committee,
				// committee id is the slot index in the epoch
				proposerId := bs.shardStatus[shardId].proposer[cid]
				shardBlock.proposer = proposerId

				// first we let the proposer to propose the block
				// check if the proposer is active
				if !bs.IsValidatorActive(proposerId) {
					continue // block skipped
				}

				// mapping the validator to the RSU
				x, y := session.SimConfig.IndexToCoord(proposerId)
				proposerRSU := (*session.RSUs)[x][y]

				// get the trust value offsets list
				startSlot, endSlot := proposerRSU.GetNextUploadSlot(), slot
				for i := startSlot; i <= endSlot; i++ {
					shardBlock.tvoList[i] = proposerRSU.GetSlotInRing(i)
				}

				// for each member in the committee,
				// start to vote for the new block
				committee := bs.GetCommitteeByCommitteeId(uint32(shardId), cid)

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
					// the good voter will high not let the bad RSU go
					case proposerIsCompromised && !validatorIsCompromised:
						switch {
						case rn < 0.8:
							shardBlock.votes[index] = false
						default:
							shardBlock.votes[index] = true
						}
					case !proposerIsCompromised && !validatorIsCompromised:
						shardBlock.votes[index] = true
					// bad validator will camouflage itself by voting for good RSU
					case !proposerIsCompromised && validatorIsCompromised:
						switch {
						case rn < 0.8:
							shardBlock.votes[index] = true
						default:
							shardBlock.votes[index] = false
						}

					}
				} // voting

				// check the voting stakes to decide whether the block should pass or not
				totalStake, gainedStake := float32(0), float32(0)
				for index, approved := range shardBlock.votes {
					vid := committee[index]
					if !bs.IsValidatorActive(uint32(index)) {
						continue
					}

					totalStake += bs.validators[int(vid)].effectiveStake
					if approved {
						gainedStake += bs.validators[int(vid)].effectiveStake
					}
				}
				// if the block gained enough stakes
				if gainedStake > totalStake*2/3 {
					shardBlock.skipped = false
				}

			} // shard block
		} // each experiments

	}
}

func (session *IBDTMSession) dialDTModule(ctx context.Context, epoch uint32) {
	select {
	case <-ctx.Done():
		logutil.LoggerList["ib-dtm"].Fatalf("[dialDTModule] context canceled")
	default:
		for expName, data := range session.TrustValueStorage {
			res := shared.IBDTM2DTMCommunication{
				Epoch:          epoch,
				ExpName:        expName,
				TrustValueList: data,
			}
			session.ChanDTM <- &res
		}
		session.ChanDTM <- true
	}
}

func (session *IBDTMSession) WaitForSimulator(ctx context.Context) error {
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
		logutil.LoggerList["dtm"].Debugf("[WaitForSimulator] init finished!")
		return nil
	}

}