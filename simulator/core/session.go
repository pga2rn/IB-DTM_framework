package core

import (
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/shared/randutil"
	"github.com/pga2rn/ib-dtm_framework/shared/timefactor"
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
	Ticker timeutil.Ticker
	// epoch and slot stored in session should only be used when gathering reports
	Epoch uint64
	Slot  uint64

	// current status
	// vehicle
	ActiveVehiclesNum         int
	ActiveVehiclesBitMap      *bitmap.Threadsafe
	MisbehaviorVehicleBitMap  *bitmap.Threadsafe
	MisbehaviorVehiclePortion float32
	// RSU
	CompromisedRSUPortion float32
	// store the ID(index) of compromised RSU of this slot
	CompromisedRSUBitMap *bitmap.Threadsafe
	// a complete list that stores every vehicle's trust value
	// TODO: not yet utilize the following 2 fields for trust value storage
	//AccurateTrustValueList []float32 // without bias
	//BiasedTrustValueList   []float32 // with bias

	// a list of all vehicles in the map
	Vehicles []*vehicle.Vehicle
	// a 2d array store the RSU data structure
	// aligned with the map structure
	RSUs [][]*dtm.RSU

	// a random generator, for determined random
	R *rand.Rand
}

// construct a simulationsession object
func PrepareSimulationSession(cfg *config.Config) *SimulationSession {
	sim := &SimulationSession{}
	sim.Config = cfg

	// init time factor
	timefactor.InitTimeFactor(cfg.SlotsPerEpoch)

	// init map
	m := simmap.CreateMap(cfg)
	sim.Map = m

	// init each data fields
	sim.ActiveVehiclesNum = 0
	sim.ActiveVehiclesBitMap = bitmap.NewTS(int(sim.Config.VehicleNumMax))
	sim.MisbehaviorVehicleBitMap = bitmap.NewTS(int(sim.Config.VehicleNumMax))
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

	sim.CompromisedRSUBitMap = bitmap.NewTS(100) // all 0 bits
	sim.CompromisedRSUPortion = 0

	// ticker
	sim.Ticker = timeutil.GetSlotTicker(cfg.Genesis, cfg.SecondsPerSlot)

	// random
	sim.R = randutil.InitRand(123)

	return sim
}

// a little helper function to convert index to coord
func (sim *SimulationSession) IndexToCoord(index int) (int, int) {
	return index / int(sim.Config.YLen), index % int(sim.Config.YLen)
}

// coord to index
func (sim *SimulationSession) CoordToIndex(x, y int) int {
	return x*int(sim.Config.YLen) + y
}
