package config

import (
	"github.com/pga2rn/ib-dtm_framework/rpc/pb"
	"github.com/pga2rn/ib-dtm_framework/shared/timefactor"
)

type ExperimentType uint32

type ExperimentConfig struct {
	Name        string
	Description string

	Type pb.ExperimentType // baseline proposal

	// has compromisedRSU or not
	CompromisedRSUFlag bool
	// apply time factor or not
	TimeFactorFlag bool
	TimeFactorType int
	// epochs traceback
	TrustValueOffsetsTraceBackEpochs int
}

func InitProposalExperimentConfigList() []*ExperimentConfig {
	m := *InitExperimentConfig()
	res := make([]*ExperimentConfig, len(m))

	count := 0
	for _, value := range m {
		if value.Type == pb.ExperimentType_PROPOSAL {
			res[count] = value
			count++
		}
	}

	return res[:count]
}

// return a map of experiment config
func InitExperimentConfig() *map[string]*ExperimentConfig {
	m := make(map[string]*ExperimentConfig)

	// baseline 0
	m["Baseline0"] = &ExperimentConfig{
		Name:                             "Baseline0",
		Description:                      "base line setup 0",
		Type:                             pb.ExperimentType_BASELINE,
		CompromisedRSUFlag:               false,
		TimeFactorFlag:                   false,
		TimeFactorType:                   -1,
		TrustValueOffsetsTraceBackEpochs: 3,
	}

	// baseline 1
	m["Baseline1"] = &ExperimentConfig{
		Name:                             "Baseline1",
		Description:                      "base line setup 1",
		Type:                             pb.ExperimentType_BASELINE,
		CompromisedRSUFlag:               true,
		TimeFactorFlag:                   false,
		TimeFactorType:                   -1,
		TrustValueOffsetsTraceBackEpochs: 3,
	}

	// proposal 0
	m["Proposal0"] = &ExperimentConfig{
		Name:                             "Proposal0",
		Description:                      "proposal 0",
		Type:                             pb.ExperimentType_PROPOSAL,
		CompromisedRSUFlag:               false,
		TimeFactorFlag:                   true,
		TimeFactorType:                   timefactor.Power,
		TrustValueOffsetsTraceBackEpochs: 3,
	}

	//proposal 1
	m["Proposal1"] = &ExperimentConfig{
		Name:                             "Proposal1",
		Description:                      "proposal 1",
		Type:                             pb.ExperimentType_PROPOSAL,
		CompromisedRSUFlag:               true,
		TimeFactorFlag:                   true,
		TimeFactorType:                   timefactor.Power,
		TrustValueOffsetsTraceBackEpochs: 3,
	}

	return &m
}
