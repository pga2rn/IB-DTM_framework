package config

import (
	"math"
	"time"
)

type IBDTMConfig struct {
	VehiclesNum        int
	ValidatorsNum      int
	UploaderCoverRatio float32 // portion of rsu that can upload within an epoch

	Genesis        time.Time
	SlotsPerEpoch  uint32
	SecondsPerSlot uint32 // in seconds
	ShardNum       int

	CommitteeSize int
	CommitteeNum  int

	InitialEffectiveStake    float32
	InitialITStake           float32
	EffectiveStakeLowerBound float32
	EffectiveStakeUpperBound float32
	BaseReward               float32
	PenaltyFactor            float32

	// miscs
	SlashingsLimit       int
	WhistleBlowingsLimit int
}

func GenIBDTMConfig(simCfg *SimConfig) *IBDTMConfig {
	res := &IBDTMConfig{
		VehiclesNum:        simCfg.VehicleNumMax,
		ValidatorsNum:      simCfg.RSUNum,
		SlotsPerEpoch:      simCfg.SlotsPerEpoch,
		Genesis:            simCfg.Genesis,
		SecondsPerSlot:     simCfg.SecondsPerSlot,
		UploaderCoverRatio: 0.5, // ratio of validators that can upload within a slot
	}

	res.CommitteeNum = int(res.SlotsPerEpoch) // one to one map to slots
	res.CommitteeSize = res.ValidatorsNum / res.CommitteeNum
	// the number of shards to allow UploaderCoverRatio portion of rsu upload blocks within an epoch
	res.ShardNum = int(float32(res.ValidatorsNum)*res.UploaderCoverRatio) / int(res.SlotsPerEpoch)

	res.InitialEffectiveStake = float32(simCfg.VehicleNumMax) * 1.5 / float32(simCfg.RSUNum)
	res.InitialITStake = float32(simCfg.VehicleNumMax * int(simCfg.SlotsPerEpoch) / simCfg.RSUNum)
	res.EffectiveStakeLowerBound = 0.5 * res.InitialEffectiveStake
	res.EffectiveStakeUpperBound = 1.5 * res.InitialEffectiveStake

	res.BaseReward =
		res.EffectiveStakeUpperBound * res.UploaderCoverRatio /
			float32(math.Sqrt(float64(simCfg.VehicleNumMax)))
	res.PenaltyFactor = 3

	res.SlashingsLimit = 4
	res.WhistleBlowingsLimit = 8

	return res
}
