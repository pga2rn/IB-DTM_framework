package config

import "time"

type IBDTMConfig struct {
	VehiclesNum              int
	ValidatorsNum            int
	UploaderCoverRatioFactor int

	Genesis        time.Time
	SlotsPerEpoch  uint32
	SecondsPerSlot uint32 // in seconds

	ActiveValidatorUpperBound int
	CommitteeSize             int
	CommitteeNum              int
}

func GenIBDTMConfig(simCfg *SimConfig) *IBDTMConfig {
	res := &IBDTMConfig{
		VehiclesNum:              simCfg.VehicleNumMax,
		ValidatorsNum:            simCfg.RSUNum,
		SlotsPerEpoch:            simCfg.SlotsPerEpoch,
		Genesis:                  simCfg.Genesis,
		SecondsPerSlot:           simCfg.SecondsPerSlot,
		UploaderCoverRatioFactor: 2, // ratio of validators that can upload within a slot
	}

	res.ActiveValidatorUpperBound = simCfg.RSUNum * 4 / 5
	res.CommitteeSize = res.ActiveValidatorUpperBound / int(res.SlotsPerEpoch)
	res.CommitteeNum = res.CommitteeSize / 2

	return res
}
