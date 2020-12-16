package core

import (
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
	"github.com/pga2rn/ib-dtm_framework/simulator/config"
	"github.com/pga2rn/ib-dtm_framework/simulator/dtm"
	"github.com/pga2rn/ib-dtm_framework/simulator/sim-map"
	"github.com/pga2rn/ib-dtm_framework/simulator/vehicle"
	"math/rand"
)

// struct that store the status of a simulation session
type SimulationSession struct {
	// config of the current simulation session
	Config *config.Config

	// pointer to the map
	Map *simmap.Map

	// time
	Ticker     timeutil.Ticker
	Epoch	uint64
	Slot uint64

	// current status
	ActiveVehiclesNum uint64
	ActiveVehiclesBitMap bitmap.Bitmap
	CompromisedRSUPortion float32
	// store the ID(index) of compromised RSU of this slot
	CompromisedRSUBitMap bitmap.Bitmap
	// a complete list that stores every vehicle's trust value
	AccurateTrustValueList []float32 // without bias
	BiasedTrustValueList []float32 // with bias

	// a list of all vehicles in the map
	Vehicles []*vehicle.Vehicle
	RSUs []*dtm.RSU

	// a random generater, for determined random
	R *rand.Rand
}

// construct a simulationsession object
func PrepareSimulationSession(cfg *config.Config) *SimulationSession{
	sim := &SimulationSession{}
	sim.Config = cfg

	// init map
	m := simmap.CreateMap(cfg)
	sim.Map = m

	// init each data fields
	sim.ActiveVehiclesNum = 0
	sim.ActiveVehiclesBitMap = bitmap.New(int(sim.Config.VehicleNumMax))
	sim.Vehicles = make([]*vehicle.Vehicle, cfg.VehicleNumMax)
	sim.RSUs = make([]*dtm.RSU, cfg.XLen * cfg.YLen)
	sim.CompromisedRSUBitMap = bitmap.New(100) // all 0 bits
	sim.CompromisedRSUPortion = 0

	// ticker
	sim.Ticker = timeutil.GetSlotTicker(cfg.Genesis, cfg.SecondsPerSlot)

	// random
	sim.R = rand.New(rand.NewSource(123))

	return sim
}