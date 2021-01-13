package ib_dtm

import (
	"context"
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/randutil"
	"sync"
)

type ShardStatus struct {
	Epoch          uint32
	Id             uint32   // shardId
	shuffledIdList []uint32 // [shuffledIndex]validatorId
	proposer       []uint32 // [committeeId]proposerId
}

type BeaconStatus struct {
	validatorMu      sync.Mutex
	slashingMu       sync.Mutex
	whistleBlowingMu sync.Mutex

	SimConfig   *config.SimConfig
	IBDTMConfig *config.IBDTMConfig
	ExpConfig   *config.ExperimentConfig

	Epoch      uint32
	Blockchain *BlockchainRoot

	activeValidators         map[uint32]*Validator
	inactivedValidatorBitMap bitmap.Bitmap
	validators               []*Validator
	slashings                map[uint32]bool
	whistleBlowings          map[uint32]int

	// committees
	shardStatus         []*ShardStatus
	TotalProposerBitMap *bitmap.Threadsafe

	// random source
	R *randutil.RandUtil
}

func InitBeaconStatus(simCfg *config.SimConfig, ibdtmConfig *config.IBDTMConfig, exp *config.ExperimentConfig, blockchain *BlockchainRoot) *BeaconStatus {
	res := &BeaconStatus{
		validatorMu:      sync.Mutex{},
		slashingMu:       sync.Mutex{},
		whistleBlowingMu: sync.Mutex{},

		SimConfig:   simCfg,
		IBDTMConfig: ibdtmConfig,
		ExpConfig:   exp,
		Blockchain:  blockchain,
	}

	// init the data structure
	res.activeValidators = make(map[uint32]*Validator)
	res.validators = make([]*Validator, simCfg.RSUNum)
	res.inactivedValidatorBitMap = bitmap.New(simCfg.RSUNum)
	res.slashings = make(map[uint32]bool)
	res.whistleBlowings = make(map[uint32]int)

	// init validator instances for every RSU
	for i := 0; i < simCfg.RSUNum; i++ {
		// register all RSUs as validator
		res.validators[i] = InitValidator(uint32(i), ibdtmConfig.InitialEffectiveStake, exp.TrustValueOffsetsTraceBackEpochs)
		// all validators are active right now
		res.activeValidators[uint32(i)] = res.validators[i]
	}

	// init every shard status storage
	res.shardStatus = make([]*ShardStatus, ibdtmConfig.ShardNum)
	for i := 0; i < ibdtmConfig.ShardNum; i++ {
		res.shardStatus[i] = &ShardStatus{
			Id:    uint32(i),
			Epoch: 0,
		}
	}
	res.TotalProposerBitMap = bitmap.NewTS(res.SimConfig.RSUNum)

	// random source
	res.R = randutil.InitRand(123)

	return res
}

func (bs *BeaconStatus) genAssignment(ctx context.Context, shardId, epoch uint32) {
	select {
	case <-ctx.Done():
		logutil.LoggerList["ib-dtm"].Fatalf("[genAssignment] context canceled")
	default:
		shardStatus := bs.shardStatus[shardId]
		if shardStatus.Epoch != epoch && epoch != 0 {
			logutil.LoggerList["ib-dtm"].Fatalf("[genAssignment] epoch async, status e %v, epoch %v", shardStatus.Epoch)
		}

		// re-generate shuffled list
		shardStatus.shuffledIdList = bs.R.PermUint32(bs.IBDTMConfig.ValidatorsNum)
		// reset proposer list
		shardStatus.proposer = make([]uint32, bs.IBDTMConfig.CommitteeSize)
		// elect proposer for each committee
		for i := 0; i < bs.IBDTMConfig.CommitteeNum; i++ {
			for {
				index := bs.R.Intn(bs.IBDTMConfig.CommitteeSize) // index inside the committee
				// proposer: [committeeId]proposerId
				proposerId := shardStatus.shuffledIdList[i*bs.IBDTMConfig.CommitteeSize+index]

				//logutil.LoggerList["ib-dtm"].Debugf("[genAssignment] p %v", proposerId)

				// prevent the proposer to proposer multiple time in different shard
				if !bs.IsValidatorActive(proposerId) {
					continue
				} else {
					shardStatus.proposer[i] = proposerId
					break
				}
			}
		}
	}
}

func (bs *BeaconStatus) GetCommitteeIdAndIndexByValidatorId(shardId, vid uint32) (cid uint32, index uint32) {
	position := 0
	for i, validator := range bs.shardStatus[shardId].shuffledIdList {
		if validator == vid {
			position = i
			break
		}
	}
	return uint32(position / bs.IBDTMConfig.CommitteeSize), uint32(position % bs.IBDTMConfig.CommitteeSize)
}

func (bs *BeaconStatus) GetCommitteeByValidatorId(shardId, vid uint32) []uint32 {
	cid, _ := bs.GetCommitteeIdAndIndexByValidatorId(shardId, vid)
	return bs.GetCommitteeByCommitteeId(shardId, cid)
}

func (bs *BeaconStatus) GetCommitteeByCommitteeId(shardId, cid uint32) []uint32 {
	id, shard := int(cid), bs.shardStatus[shardId]
	return shard.shuffledIdList[id*bs.IBDTMConfig.CommitteeSize : (id+1)*bs.IBDTMConfig.CommitteeSize]
}

func (bs *BeaconStatus) IsValidatorActive(vid uint32) bool {
	bs.validatorMu.Lock()
	res := !bs.inactivedValidatorBitMap.Get(int(vid))
	bs.validatorMu.Unlock()
	return res
}

func (bs *BeaconStatus) InactivateValidator(vid uint32) {
	bs.validatorMu.Lock()
	delete(bs.activeValidators, vid)
	bs.inactivedValidatorBitMap.Set(int(vid), true)
	bs.validatorMu.Unlock()
}

func (bs *BeaconStatus) ActivateValidator(vid uint32) {
	bs.validatorMu.Lock()
	bs.activeValidators[vid] = bs.validators[int(vid)]
	bs.inactivedValidatorBitMap.Set(int(vid), false)
	bs.validatorMu.Unlock()
}

func (bs *BeaconStatus) GetRewardFactor(id uint32) float32 {
	//validator := bs.validators[id]
	//res := validator.itsStake / float32(bs.IBDTMConfig.VehiclesNum)
	return 0.75 // TODO: change the hard coded reward factor!
}

func (bs *BeaconStatus) UpdateShardStatus(ctx context.Context, epoch uint32) {
	logutil.LoggerList["ib-dtm"].Debugf("[UpdateShardStatus] epoch %v", epoch)
	defer logutil.LoggerList["ib-dtm"].Debugf("[UpdateShardStatus] epoch %v, done", epoch)

	// reset the bitmap
	bs.TotalProposerBitMap = bitmap.NewTS(bs.SimConfig.RSUNum)
	for shardId := range bs.shardStatus {
		bs.shardStatus[shardId] = &ShardStatus{
			Epoch: epoch,
			Id:    uint32(shardId),
		}
		bs.genAssignment(ctx, uint32(shardId), epoch)
	}
}
