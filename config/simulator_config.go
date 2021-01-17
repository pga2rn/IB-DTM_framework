// Package config defines the config for the simulation
package config

import (
	"time"
)

var Genesis time.Time

func SetGenesis(genesis time.Time) {
	Genesis = genesis
}

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
	EpochCacheLength int

	// time config
	Genesis           time.Time
	SlotsPerEpoch     uint32
	SecondsPerSlot    uint32 // in seconds
	OutOfSyncTolerant uint32 // in slots
	FinalizedDelay    uint32 // in epoch

	// rsu config
	RingLength int

	// vehicle config
}

func GenYangNetConfig() *SimConfig {
	return &SimConfig{
		Genesis:        Genesis,
		SecondsPerSlot: 1,
		SlotsPerEpoch:  16,
		RSUNum:         256,

		XLen: 16,
		YLen: 16,

		EpochCacheLength:      96,
		CompromisedRSUPortion: 0.2,

		MisbehaveVehiclePortion: 0.2,
		VehicleNumMin:           8000,
		VehicleNumMax:           8192,
	}
}

// a little helper function to convert index to coord
func (cfg *SimConfig) IndexToCoord(index uint32) (int, int) {
	return int(index) / cfg.YLen, int(index) % cfg.YLen
}

// coord to index
func (cfg *SimConfig) CoordToIndex(x, y int) uint32 {
	return uint32(x*cfg.YLen + y)
}
