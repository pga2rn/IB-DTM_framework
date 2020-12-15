package core

import (
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
	"github.com/pga2rn/ib-dtm_framework/simulator/config"
	"github.com/pga2rn/ib-dtm_framework/simulator/dtm"
	"github.com/pga2rn/ib-dtm_framework/simulator/sim-map"
	"github.com/pga2rn/ib-dtm_framework/simulator/vehicle"
	"math/rand"
	"time"
)

type Beacon struct {
	// genesis
	Genesis time.Time
	// time sync
	Epoch uint64
	Slot uint64
}

// struct that store the status of a simulation session
type SimulationSession struct {
	// config of the current simulation session
	Config *config.Config

	// pointer to the map
	Map *simmap.Map

	// time
	Ticker     timeutil.Ticker
	TimeStream Beacon

	// current status
	ActiveVehiclesNum uint64
	CompromisedRSUPortion float32
	// store the ID(index) of compromised RSU of this slot
	CompromisedRSUList []*int
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
	sim.Vehicles = make([]*vehicle.Vehicle, cfg.VehicleNumMax)
	sim.RSUs = make([]*dtm.RSU, cfg.XLen * cfg.YLen)

	// ticker
	sim.Ticker = timeutil.GetSlotTicker(cfg.Genesis, cfg.SecondsPerSlot)

	// random
	sim.R = rand.New(rand.NewSource(123))

	return sim

}

func (sim *SimulationSession) Done(){
	// terminate the ticker
	sim.Ticker.Done()
	return
}

// wait for connecting with external RSU modules
func (sim *SimulationSession) WaitForRSUInit() error {
	return nil
}

// initializing vehicles, place VehicleNumMin vehicles into the network
func (sim *SimulationSession) WaitForVehiclesInit() error {
	sim.ActiveVehiclesNum = sim.Config.VehicleNumMin

	// activate the very first ActivateVehiclesNum vehicles
	for i := 0 ; i < int(sim.ActiveVehiclesNum); i++ {
		v := &vehicle.Vehicle{}
		v.Pos = vehicle.Position {
			sim.R.Intn(int(sim.Config.XLen)),
			sim.R.Intn(int(sim.Config.YLen)),
			}
		v.VehicleStatus = vehicle.Active
		v.LastMovementDirection = vehicle.NotMove

		// register the vehicle to the session
		sim.Vehicles[i] = v

		// place the vehicle onto the map
		sim.Map.Cross
		}

	return nil
}