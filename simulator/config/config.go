// Package config defines the config for the simulation
package config

import (
	"math/rand"
	"time"
)

// a global accessable random object
var R = rand.New(rand.NewSource(time.Now().Unix()))

// Config is used to define a simulation
type Config struct {
	////// map config ///////
	XLen uint32
	YLen uint32

	////// simulation config //////
	VehicleNumMax uint64
	VehicleNumMin uint64
	RSUNum uint32
	PortionOfCompromisedRSUMax float32 // from 0 ~ 1
	PortionOfCompromisedRSUMin float32 // from 0 ~ 1

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
	cfg.RSUNum = 100

	// map config
	cfg.XLen = 10
	cfg.YLen = 10

	// sim config
	cfg.OutOfSyncTolerant = 1 // only allow 1 slot out-of-sync
	cfg.FinalizedDelay = 2 // aligned with eth2.0 setup
	cfg.PortionOfCompromisedRSUMax = 0.25
	cfg.PortionOfCompromisedRSUMin = 0.05
	cfg.VehicleNumMin = 600
	cfg.VehicleNumMax = 1000

	return cfg
}

func (cfg *Config) ConfigSetGensis(genesis time.Time) {
	cfg.Genesis = genesis
}