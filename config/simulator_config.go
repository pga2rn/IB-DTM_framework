// Package config defines the config for the simulation
package config

import (
	"time"
)

// SimConfig is used to define a simulation
type SimConfig struct {
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

	// how many previous epochs' tvos will be used to calculate tv
	TrustValueOffsetsTraceBackEpoch int

	// time config
	Genesis           time.Time
	SlotsPerEpoch     uint64
	SecondsPerSlot    uint64 // in seconds
	OutOfSyncTolerant uint64 // in slots
	FinalizedDelay    uint64 // in epoch

	// rsu config
	RingLength int

	// vehicle config
}

func GenYangNetConfig() *SimConfig {
	cfg := &SimConfig{}

	// config aligned to yang test eth2 net
	cfg.SecondsPerSlot = 6
	cfg.SlotsPerEpoch = 2
	cfg.RSUNum = 25

	// map config
	cfg.XLen = 5
	cfg.YLen = 5

	// sim config
	cfg.OutOfSyncTolerant = 1 // only allow 1 slot out-of-sync
	cfg.FinalizedDelay = 2    // aligned with eth2.0 setup
	cfg.TrustValueOffsetsTraceBackEpoch = 3

	// rsu config
	cfg.CompromisedRSUPortionMax = 0.25
	cfg.CompromisedRSUPortionMin = 0.05
	cfg.RingLength = cfg.TrustValueOffsetsTraceBackEpoch * int(cfg.SlotsPerEpoch)

	// vehicle
	cfg.MisbehaveVehiclePortionMax = 0.3
	cfg.MisbehaveVehiclePortionMin = 0.05
	cfg.VehicleNumMin = 600
	cfg.VehicleNumMax = 1000

	return cfg
}

func (cfg *SimConfig) SetGenesis(genesis time.Time) {
	cfg.Genesis = genesis
}
