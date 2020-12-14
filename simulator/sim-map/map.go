package simmap

import (
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
	"github.com/pga2rn/ib-dtm_framework/simulator/config"
	"github.com/pga2rn/ib-dtm_framework/simulator/rsu"
	"github.com/pga2rn/ib-dtm_framework/simulator/vehicle"
)

type Position struct {
	X uint32
	Y uint32
}

// each cross represents a CROSS within the map,
// which holds a RSU and 0 or more vehicles
// map looks like this:
// 0->N
// |
// v
// M
type cross struct {
	// position of the cross
	Pos Position

	// pointer to RSU
	Rsu *rsu.RSU

	// a list of vehicle that appears
	// convention: index is the ID of the vehicle
	Vehicles []*vehicle.Vehicle
}

type mapStatus struct {
	// time
	Genesis uint64
	Epoch   uint64 // current epoch
	Slot    uint64 // current slot

	// current status
	ActiveVehiclesNum uint64
	CompromisedRSUPortion float32
	// a complete list that stores every vehicle's trust value
	TrustValueList []float32
}

type Map struct {
	////// config /////
	SimConfig config.Config

	// current status of the whole map
	MapStatus mapStatus

	////// map /////
	Ticker timeutil.Ticker
	Cross []*cross
	// a list of all vehicles in the map
	// index is the ID of the vehicle
	Vehicles []*vehicle.Vehicle
}

// create a brand new map
func CreateMap(cfg config.Config) *Map {
	m := &Map{}

	// prepare the ticker
	m.Ticker = timeutil.GetSlotTicker(cfg.Genesis, cfg.SlotLen)

	// prepare the map
	m.Cross = make([]*cross, cfg.XLen * cfg.YLen)
	// assign each cross to a RSU
	for i:=0;i<len(m.Cross);i++{
		// pass
	}

	// prepare the list of vehicles (with upper bound)
	m.vehicles = make([]*vehicle.Vehicle, cfg.VehicleNumMax)
	// init every vehicle
	for i:=0;i<len(m.vehicles);i++{
		// pass
	}

	return m
}