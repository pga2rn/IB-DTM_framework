package simulator

import (
	"context"
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/rsu"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/randutil"
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
	"github.com/pga2rn/ib-dtm_framework/sim-map"
	"github.com/pga2rn/ib-dtm_framework/vehicle"
	"sync"
)

// struct that store the status of a simulation session
type SimulationSession struct {
	// config of the current simulation session
	Config *config.SimConfig

	// pointer to the map
	Map *simmap.Map

	// channel for inter-module-communication
	ChanDTM   chan interface{}
	ChanIBDTM chan interface{}

	// time
	Ticker timeutil.Ticker
	// epoch and slot stored in session should only be used when gathering reports
	Epoch uint32
	Slot  uint32

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
	vmu      sync.Mutex
	// a 2d array store the RSU data structure
	// aligned with the map structure
	RSUs [][]*rsu.RSU
	rmu  sync.Mutex
	// a random generator, for determined random
	R *randutil.RandUtil
}

// construct a simulationsession object
func PrepareSimulationSession(cfg *config.SimConfig, chanDTM chan interface{}, chanIBDTM chan interface{}) *SimulationSession {
	sim := &SimulationSession{}
	sim.Config = cfg

	// inter module
	sim.ChanDTM = chanDTM
	sim.ChanIBDTM = chanIBDTM

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
	sim.RSUs = make([][]*rsu.RSU, cfg.YLen)
	for x := range sim.RSUs {
		sim.RSUs[x] = make([]*rsu.RSU, cfg.XLen)
	}

	sim.CompromisedRSUBitMap = bitmap.NewTS(100) // all 0 bits
	sim.CompromisedRSUPortion = 0

	// ticker
	sim.Ticker = timeutil.GetSlotTicker(cfg.Genesis, cfg.SecondsPerSlot)

	// random
	sim.R = randutil.InitRand(123)

	return sim
}

func (sim *SimulationSession) dialIBDTMLogicModulePerSlot(ctx context.Context, slot uint32) {
	select {
	case <-ctx.Done():
		logutil.LoggerList["sim"].Fatalf("[dialIBDTMLogicModulePerSlot] context canceled")
	default:
		// signal the ib-dtm module with slot
		sim.ChanIBDTM <- slot
		// wait for the process finished
		select {
		case <-ctx.Done():
			logutil.LoggerList["sim"].Fatalf("[dialIBDTMLogicModulePerSlot] context canceled for ibdtm slot %v process", slot)
		case <-sim.ChanIBDTM:
			logutil.LoggerList["sim"].Debugf("[dialIBDTMLogicModulePerSlot] slot %v done", slot)
		}
	}
}
