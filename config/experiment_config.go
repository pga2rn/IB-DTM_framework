package config

import "github.com/pga2rn/ib-dtm_framework/shared/timefactor"

type ExperimentConfig struct {
	Name        string
	Description string

	Type int // baseline proposal

	// has compromisedRSU or not
	CompromisedRSUFlag bool
	// apply time factor or not
	TimeFactorFlag bool
	TimeFactorType int
	// epochs traceback
	TrustValueOffsetsTraceBackEpochs int
}

const (
	Baseline = iota
	Proposal
)

// return a map of experiment config
func InitExperimentConfig() *map[string]*ExperimentConfig {
	m := make(map[string]*ExperimentConfig)

	// baseline 0
	m["Baseline0"] = &ExperimentConfig{
		Name:                             "Baseline0",
		Description:                      "base line setup 0",
		Type:                             Baseline,
		CompromisedRSUFlag:               false,
		TimeFactorFlag:                   false,
		TimeFactorType:                   -1,
		TrustValueOffsetsTraceBackEpochs: 3,
	}

	// baseline 2
	m["Baseline2"] = &ExperimentConfig{
		Name:                             "Baseline2",
		Description:                      "base line setup 2",
		Type:                             Baseline,
		CompromisedRSUFlag:               true,
		TimeFactorFlag:                   false,
		TimeFactorType:                   -1,
		TrustValueOffsetsTraceBackEpochs: 3,
	}

	// proposal 0
	m["Proposal0"] = &ExperimentConfig{
		Name:                             "Proposal0",
		Description:                      "proposal 0",
		Type:                             Proposal,
		CompromisedRSUFlag:               false,
		TimeFactorFlag:                   true,
		TimeFactorType:                   timefactor.Power,
		TrustValueOffsetsTraceBackEpochs: 3,
	}

	// proposal 1
	m["Proposal1"] = &ExperimentConfig{
		Name:                             "Proposal1",
		Description:                      "proposal 1",
		Type:                             Proposal,
		CompromisedRSUFlag:               true,
		TimeFactorFlag:                   true,
		TimeFactorType:                   timefactor.Power,
		TrustValueOffsetsTraceBackEpochs: 3,
	}

	return &m
}
