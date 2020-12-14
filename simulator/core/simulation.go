package core

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
	"github.com/pga2rn/ib-dtm_framework/simulator/config"
	"github.com/pga2rn/ib-dtm_framework/simulator/dtm"
	"github.com/pga2rn/ib-dtm_framework/simulator/rsu"
	"github.com/pga2rn/ib-dtm_framework/simulator/sim-map"
	"github.com/pga2rn/ib-dtm_framework/simulator/vehicle"
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
	RealTrustValueList []float32 // without bias
	EffectiveTrustValueList []float32 // with bias

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

}

// start the simulation!
// routines are as follow:
// case 1: main process exit, simulation stop
// case 2: waiting for next slot
//		r1: update the map, move the vehicles
//		r2: generate trust value for newly moved vehicles
//		r3:	call RSU, provide trust value offsets to them and let them do the job
//		r2: calculate trust value
func run(ctx context.Context, sim *SimulationSession){
	cleanup := sim.Done
	defer cleanup()

	// init RSU
	// wait for every RSU to comes online
	// sim.WaitForRSUInit()

	// init vehicles
	// sim.PrepareVehicles()


	// start the main loop
	for {
		ctx, cancel := context.WithCancel(ctx)

		//select {
		//case <-ctx.Done():
		//	cancel()
		//	return // exit
		//
		//	case slot := <- sim.Ticker.NextSlot():
		//
		//
		//
		//}
	}


}