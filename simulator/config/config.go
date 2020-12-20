// Package config defines the config for the simulation
package config

import (
	"github.com/pga2rn/ib-dtm_framework/shared/timefactor"
	"time"
)

// Config is used to define a simulation
type Config struct {
	////// map config ///////
	XLen int
	YLen int

	////// simulation config //////
	VehicleNumMax              int
	VehicleNumMin              int
	MisbehaveVehiclePortionMax float32
	MisbehaveVehiclePortionMin float32

	RSUNum                   int
	CompromisedRSUPortionMax float32 // from 0 ~ 1
	CompromisedRSUPortionMin float32 // from 0 ~ 1

	TimeFactorType int

	// time config
	Genesis           time.Time
	SlotsPerEpoch     uint64
	SecondsPerSlot    uint64 // in seconds
	OutOfSyncTolerant uint64 // in slots
	FinalizedDelay    uint64 // in epoch

	// vehicle config
}

type RSUConfig struct {
}

func GenYangNetConfig() *Config {
	cfg := &Config{}

	// config aligned to yang test eth2 net
	cfg.SecondsPerSlot = 6
	cfg.SlotsPerEpoch = 6
	cfg.RSUNum = 400

	// map config
	cfg.XLen = 20
	cfg.YLen = 20

	// sim config
	cfg.OutOfSyncTolerant = 1 // only allow 1 slot out-of-sync
	cfg.FinalizedDelay = 2    // aligned with eth2.0 setup
	cfg.TimeFactorType = timefactor.Power

	// rsu config
	cfg.CompromisedRSUPortionMax = 0.25
	cfg.CompromisedRSUPortionMin = 0.05

	// vehicle
	cfg.MisbehaveVehiclePortionMax = 0.3
	cfg.MisbehaveVehiclePortionMin = 0.05
	cfg.VehicleNumMin = 600
	cfg.VehicleNumMax = 1000

	return cfg
}

func (cfg *Config) SetGenesis(genesis time.Time) {
	cfg.Genesis = genesis
}
