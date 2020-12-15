package core

import (
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
	"github.com/pga2rn/ib-dtm_framework/simulator/config"
	"github.com/pga2rn/ib-dtm_framework/simulator/dtm"
	"github.com/pga2rn/ib-dtm_framework/simulator/sim-map"
	"github.com/pga2rn/ib-dtm_framework/simulator/vehicle"
	"github.com/sirupsen/logrus"
	"time"
)

var log = logrus.New()

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
}

// construct a simulationsession object
func PrepareSimulationSession(
	cfg *config.Config,
	m *simmap.Map, ) *SimulationSession{

	sim := &SimulationSession{}

	sim.Config = cfg
	sim.Map = m

	// init each data fields
	sim.ActiveVehiclesNum = 0
	sim.Vehicles = make([]*vehicle.Vehicle, cfg.VehicleNumMax)
	sim.RSUs = make([]*dtm.RSU, cfg.XLen * cfg.YLen)

	// ticker
	sim.Ticker = timeutil.GetSlotTicker(cfg.Genesis, cfg.SlotLen)

	return sim

}

func (sim *SimulationSession) Done(){
	return
}

// wait for connecting with external RSU modules
func (sim *SimulationSession) WaitForRSUInit() error {
	return nil
}

// initializing vehicles, place VehicleNumMin vehicles into the network
func (sim *SimulationSession) WaitForInitingVehicles() error {
	sim.ActiveVehiclesNum = sim.Config.VehicleNumMin
	for i := 0 ; i < int(sim.ActiveVehiclesNum); i++ {
		// init the vehicles here
	}

	return nil
}