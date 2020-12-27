package simulator

import (
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/dtm"
	"github.com/pga2rn/ib-dtm_framework/shared/randutil"
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
	"github.com/pga2rn/ib-dtm_framework/sim-map"
	"github.com/pga2rn/ib-dtm_framework/vehicle"
	"sync"
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
	TrustValueList       *sync.Map // without bias
	BiasedTrustValueList *sync.Map // with bias

	// a list of all vehicles in the map
	Vehicles []*vehicle.Vehicle
	vmu	sync.Mutex
	// a 2d array store the RSU data structure
	// aligned with the map structure
	RSUs [][]*dtm.RSU
	rmu sync.Mutex
	// a random generator, for determined random
	R *randutil.RandUtil
}

// construct a simulationsession object
func PrepareSimulationSession(cfg *config.Config) *SimulationSession {
	sim := &SimulationSession{}
	sim.Config = cfg

	// init map
	m := simmap.CreateMap(cfg)
	sim.Map = m

	// init mutex
	sim.vmu = sync.Mutex{}
	sim.rmu = sync.Mutex{}

	// init each data fields
	sim.ActiveVehiclesNum = 0
	sim.ActiveVehiclesBitMap = bitmap.NewTS(int(sim.Config.VehicleNumMax))
	sim.MisbehaviorVehicleBitMap = bitmap.NewTS(int(sim.Config.VehicleNumMax))
	sim.Vehicles = make([]*vehicle.Vehicle, cfg.VehicleNumMax)
	sim.RSUs = make([][]*dtm.RSU, cfg.YLen)
	for x := range sim.RSUs {
		sim.RSUs[x] = make([]*dtm.RSU, cfg.XLen)
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
