package ib_dtm

import (
	"context"
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
	Blockchain map[string]*BlockchainHead

	// pointer to the vehicles and RSU
	// I don't know if it is a good idea to use mutex via pointer
	RSUs *[][]*rsu.RSU
	rmu  *sync.Mutex

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
	simCfg *config.SimConfig, expCfgList []*config.ExperimentConfig,
	sim, dtm chan interface{}) *IBDTMSession {

	session := &IBDTMSession{}
	session.SimConfig = simCfg
	session.ExpConfigList = expCfgList

	session.ChanDTM, session.ChanSim = dtm, sim

	// init storage area
	for _, exp := range expCfgList {
		if exp.Type == pb.ExperimentType_PROPOSAL {
			session.Blockchain[exp.Name] = InitBlockchain()
		}
	}

	// prepare the ticker
	session.Ticker = timeutil.GetSlotTicker(simCfg.Genesis, simCfg.SecondsPerSlot)

	return session
}

func (session *IBDTMSession) processEpoch(ctx context.Context, epoch uint32) {
	select {
	case <-ctx.Done():
		logutil.LoggerList["ib-dtm"].Fatalf("[processEpoch] context canceled")
	default:
		switch epoch {
		case uint32(0):
			// generate the assignment
			session.BeaconStatus.PrepareForNextEpoch(0)
		default:
			// generate trust value
			session.genTrustValue(ctx, epoch)

			// pass the results to the dtm module
			session.dialDTModule(ctx, epoch)

			// init for the next epoch
			session.processReward(ctx)
			session.processPenalty(ctx)

			// update the beacon status
			session.BeaconStatus.PrepareForNextEpoch(epoch)
		}
	}
}

func (session *IBDTMSession) ProcessSlot(ctx context.Context, slot uint32) {
	select {
	case <-ctx.Done():
		logutil.LoggerList["ib-dtm"].Fatalf("[processSlo] context canceled")
	default:
		for _, exp := range session.ExpConfigList {
			bs := session.BeaconStatus[exp.Name]

			// init beaconblock
			blockchain := session.Blockchain[exp.Name]
			// TODO: include ib-dtm config in initBlockChain
			beaconBlock, err := blockchain.InitBlockchainBlock(slot, session.SimConfig)
			if err != nil {
				logutil.LoggerList["ib-dtm"].Fatalf("[processSlot] blockchain init failed, %v", err)
			}
			// init shard blocks
			for i := 0; i < session.IBDTMConfig.CommitteeSize; i++ {
				beaconBlock.shards[i] = &BlockchainShard{
					skipped: true,
					slot:    slot,
				}
			}

			// iterate all RSU
			for i := 0; i < session.SimConfig.RSUNum; i++ {
				x, y := session.SimConfig.IndexToCoord(uint32(i))

				// the validator has exited
				if !bs.IsValidatorActive(uint32(i)) {
					continue
				}

				r := (*session.RSUs)[x][y]

				if err, role := session.BeaconStatus.GetRole(r.Id); role == Proposer {
					continue
				} else if err != nil {
					logutil.LoggerList["ib-dtm"].Fatalf("[ProcessSlot] failed to get role for rsu %v", r.Id)
				}

				// for every experiment
				for _, exp := range session.ExpConfigList {
					// get the blockchain
					block := session.Blockchain[exp.Name].GetHeadBlock()

					// if attestor,
					// the attestors will record the evils done by the proposer if any,
					// and records it, and will propose in their proposer slot.
					// executeVoting(id, block)
				}
			} // iterate RSU

			// check the vote balance
			// iterate through proposer
			bs, blockchain := session.BeaconStatus[exp.Name], session.Blockchain[exp.Name]
			for _, rid := range bs.proposer {
				x, y := session.SimConfig.IndexToCoord(rid)
				r := (*session.RSUs)[x][y]

				shardIndex, err := session.BeaconStatus[exp.Name].GetCommitteeId(r.Id)
				if err != nil {
					return
				}

				totalBalance, gainedBalance := float32(0), float32(0)

				block := blockchain.GetHeadBlock().shards[shardIndex]
				for voter, vote := range *block.votes {
					if vote {
						gainedBalance += (*bs.Validators)[voter].EffectiveStake
					}
					totalBalance += (*bs.Validators)[voter].EffectiveStake
				}
				if gainedBalance > totalBalance*2/3 {
					block.skipped = false
				}
			} // iterate through proposer
		} // for each experiments

	} // select
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
