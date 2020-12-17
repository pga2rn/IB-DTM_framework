package core

import (
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/shared/randutil"
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
	// vehicle
	ActiveVehiclesNum uint64
	ActiveVehiclesBitMap bitmap.Bitmap
	MisbehaviorVehicleBitMap bitmap.Bitmap
	// RSU
	CompromisedRSUPortion float32
	// store the ID(index) of compromised RSU of this slot
	CompromisedRSUBitMap bitmap.Bitmap
	// a complete list that stores every vehicle's trust value
	AccurateTrustValueList []float32 // without bias
	BiasedTrustValueList []float32 // with bias

	// a list of all vehicles in the map
	Vehicles []*vehicle.Vehicle
	// a 2d array store the RSU data structure
	// aligned with the map structure
	RSUs [][]*dtm.RSU

	// a random generator, for determined random
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
	sim.MisbehaviorVehicleBitMap = bitmap.New(int(sim.Config.VehicleNumMax))
	sim.Vehicles = make([]*vehicle.Vehicle, cfg.VehicleNumMax)

	sim.RSUs = make([][]*dtm.RSU, cfg.YLen)
	for x := range sim.RSUs {
		sim.RSUs[x] = make([]*dtm.RSU, cfg.XLen)
		// init every RSU data structure
		for y := 0; y < int(cfg.XLen); y++ {
			r := dtm.RSU{}
			sim.RSUs[x][y] = &r
		}
	}

	sim.CompromisedRSUBitMap = bitmap.New(100) // all 0 bits
	sim.CompromisedRSUPortion = 0

	// ticker
	sim.Ticker = timeutil.GetSlotTicker(cfg.Genesis, cfg.SecondsPerSlot)

	// random
	sim.R = randutil.InitRand(123)

	return sim
}