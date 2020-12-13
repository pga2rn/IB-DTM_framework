// Package config defines the config for the simulation
package config

import "time"

// Config is used to define a simulation
type Config struct {
	////// map config ///////
	XLen uint32
	YLen uint32

	////// simulation config //////
	VehicleNumMax uint32
	VehicleNumMin uint32
	RSUNum uint32
	PortionOfCompromisedRSUMax float32 // from 0 ~ 1
	PortionOfCompromisedRSUMin float32 // from 0 ~ 1

	// time config
	Genesis time.Time
	SlotsPerEpoch uint64
	SlotLen uint64 // in seconds
}

func GenYangNetConfig() *Config {
	cfg := &Config{}

	// config aligned to yang test eth2 net
	cfg.slotLen = 6
	cfg.slotsPerEpoch = 6
	cfg.rsuNum = 100

	// map config
	cfg.xLen = 10
	cfg.yLen = 10

	return cfg
}

func ConfigVehicleNum(cfg *Config, min int, max int) {
	cfg.vehicleNumMin = 600
	cfg.vehicleNumMax = 1000
}

func ConfigGensis(cfg *Config, genesis uint64) {
	cfg.genesis = genesis
}