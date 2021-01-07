// Package config defines the config for the simulation
package config

import (
	"github.com/sirupsen/logrus"
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
	MisbehaveVehiclePortion    float32

	RSUNum                   int
	CompromisedRSUPortionMax float32 // from 0 ~ 1
	CompromisedRSUPortion    float32 // from 0 ~ 1

	// how many previous epochs' tvos will be used to calculate tv
	TrustValueOffsetsTraceBackEpoch int

	// time config
	Genesis           time.Time
	SlotsPerEpoch     uint32
	SecondsPerSlot    uint32 // in seconds
	OutOfSyncTolerant uint32 // in slots
	FinalizedDelay    uint32 // in epoch

	// rsu config
	RingLength int
	Loglevel   logrus.Level

	// vehicle config
}

func GenYangNetConfig() *SimConfig {
	cfg := &SimConfig{}
	cfg.Loglevel = logrus.InfoLevel

	// config aligned to yang test eth2 net
	cfg.SecondsPerSlot = 1
	cfg.SlotsPerEpoch = 16
	cfg.RSUNum = 256

	// map config
	cfg.XLen = 16
	cfg.YLen = 16

	// sim config
	cfg.TrustValueOffsetsTraceBackEpoch = 3

	// rsu config
	cfg.CompromisedRSUPortion = 0.2
	cfg.RingLength = cfg.TrustValueOffsetsTraceBackEpoch * int(cfg.SlotsPerEpoch)

	// vehicle
	cfg.MisbehaveVehiclePortion = 0.2
	cfg.VehicleNumMin = 3467
	cfg.VehicleNumMax = 4096

	return cfg
}

func (cfg *SimConfig) SetGenesis(genesis time.Time) {
	cfg.Genesis = genesis
}
