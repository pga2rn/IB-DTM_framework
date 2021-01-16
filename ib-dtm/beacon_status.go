package ib_dtm

import (
	"context"
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/randutil"
	"math"
	"sort"
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
	validators               ValidatorPointerList
	slashings                map[uint32]bool
	whistleBlowings          map[uint32]int

	// committees
	shardStatus []*ShardStatus

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
	res.validators = ValidatorPointerList{
		make(Validators, simCfg.RSUNum),
	}
	res.inactivedValidatorBitMap = bitmap.New(simCfg.RSUNum)
	res.slashings = make(map[uint32]bool)
	res.whistleBlowings = make(map[uint32]int)

	// init validator instances for every RSU
	for i := 0; i < simCfg.RSUNum; i++ {
		// register all RSUs as validator
		res.validators.Validators[i] = InitValidator(uint32(i), ibdtmConfig.InitialEffectiveStake, exp.TrustValueOffsetsTraceBackEpochs)
		// all validators are active right now
		res.activeValidators[uint32(i)] = res.validators.Validators[i]
	}

	// init every shard status storage
	res.shardStatus = make([]*ShardStatus, ibdtmConfig.ShardNum)
	for i := 0; i < ibdtmConfig.ShardNum; i++ {
		res.shardStatus[i] = &ShardStatus{
			Id:    uint32(i),
			Epoch: 0,
		}
	}

	// random source
	res.R = randutil.InitRand(123)

	return res
}

// separate proposer and committee(proposer no need to be the member of committee)
func (bs *BeaconStatus) genAssignment(ctx context.Context, shardId, epoch uint32) {
	//logutil.GetLogger(PackageName).Debugf("[genAssignment] epoch %v", epoch)
	//defer logutil.GetLogger(PackageName).Debugf("[genAssignment] epoch %v done", epoch)

	select {
	case <-ctx.Done():
		logutil.GetLogger(PackageName).Fatalf("[genAssignment] context canceled")
	default:
		shardStatus := bs.shardStatus[shardId]
		if shardStatus.Epoch != epoch && epoch != 0 {
			logutil.GetLogger(PackageName).Fatalf("[genAssignment] epoch async, status e %v, epoch %v", shardStatus.Epoch)
		}

		// re-generate shuffled list
		shardStatus.shuffledIdList = bs.R.PermUint32(bs.IBDTMConfig.ValidatorsNum)
		// reset proposer list
		shardStatus.proposer = make([]uint32, bs.IBDTMConfig.CommitteeSize)

		// take out the first (ValidatorNum * CoverRatio)th validators with most stakes
		tmpVlist := make(Validators, len(bs.validators.Validators))
		copy(tmpVlist, bs.validators.Validators)
		// sort the tmp list based on the amount of stake
		sort.Sort(ValidatorPointerList{tmpVlist})

		//
		for i := 0; i < bs.IBDTMConfig.CommitteeNum; i++ {
			for {
				// randomly pick a proposer from the first 2/3 validators with most stakes
				rn := bs.R.Intn(len(tmpVlist) * 2 / 3)

				if !bs.IsValidatorActive(tmpVlist[rn].Id) {
					continue
				} else {
					shardStatus.proposer[i] = tmpVlist[rn].Id
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
	bs.activeValidators[vid] = bs.validators.Validators[int(vid)]
	bs.inactivedValidatorBitMap.Set(int(vid), false)
	bs.validatorMu.Unlock()
}

func (bs *BeaconStatus) GetRewardFactor(id uint32) float32 {
	validator := bs.validators.Validators[id]
	res := float32(math.Sqrt(float64(validator.itsStake.GetAmount() / float32(bs.IBDTMConfig.VehiclesNum))))
	return res
}

func (bs *BeaconStatus) UpdateShardStatus(ctx context.Context, epoch uint32) {
	logutil.GetLogger(PackageName).Debugf("[UpdateShardStatus] epoch %v", epoch)
	defer logutil.GetLogger(PackageName).Debugf("[UpdateShardStatus] epoch %v, done", epoch)

	wg := sync.WaitGroup{}
	for shardId := range bs.shardStatus {
		bs.shardStatus[shardId] = &ShardStatus{
			Epoch: epoch,
			Id:    uint32(shardId),
		}

		wg.Add(1)
		go func(shardId uint32) { // spawn go routines for assignment generation
			bs.genAssignment(ctx, shardId, epoch)
			wg.Done()
		}(uint32(shardId))
	}
	wg.Wait()
}
