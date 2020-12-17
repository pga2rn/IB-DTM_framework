// Package config defines the config for the simulation
package config

import (
	"time"
)

// Config is used to define a simulation
type Config struct {
	////// map config ///////
	XLen uint32
	YLen uint32

	////// simulation config //////
	VehicleNumMax            int
	VehicleNumMin            int
	MisbehaveVehiclePortionMax float32
	MisbehaveVehiclePortionMin float32

	RSUNum                   int
	CompromisedRSUPortionMax float32 // from 0 ~ 1
	CompromisedRSUPortionMin float32 // from 0 ~ 1



	// time config
	Genesis           time.Time
	SlotsPerEpoch     uint64
	SecondsPerSlot           uint64 // in seconds
	OutOfSyncTolerant uint64 // in slots
	FinalizedDelay	uint64 // in epoch

	// vehicle config
}

type RSUConfig struct {
	
}

func GenYangNetConfig() *Config {
	cfg := &Config{}

	// config aligned to yang test eth2 net
	cfg.SecondsPerSlot = 6
	cfg.SlotsPerEpoch = 6
	cfg.RSUNum = 25

	// map config
	cfg.XLen = 5
	cfg.YLen = 5

	// sim config
	cfg.OutOfSyncTolerant = 1 // only allow 1 slot out-of-sync
	cfg.FinalizedDelay = 2 // aligned with eth2.0 setup

	// rsu config
	cfg.CompromisedRSUPortionMax = 0.25
	cfg.CompromisedRSUPortionMin = 0.05

	// vehicle
	cfg.MisbehaveVehiclePortionMax = 0.3
	cfg.MisbehaveVehiclePortionMin = 0.05
	cfg.VehicleNumMin = 12
	cfg.VehicleNumMax = 20

	return cfg
}

func (cfg *Config) SetGenesis(genesis time.Time) {
	cfg.Genesis = genesis
}