package simmap

import (
	"github.com/pga2rn/ib-dtm_framework/simulator/config"
	"github.com/pga2rn/ib-dtm_framework/simulator/core"
	"github.com/pga2rn/ib-dtm_framework/simulator/vehicle"
)

type Position struct {
	X int
	Y int
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

	// a list of vehicle that appears
	Vehicles []*vehicle.Vehicle
}

type Map struct {
	////// map /////
	Cross []*cross

	SimStatus *core.SimulationSession
}

// create a brand new map
func CreateMap(cfg config.Config) *Map {
	m := &Map{}

	// prepare the map
	m.Cross = make([]*cross, cfg.XLen * cfg.YLen)

	return m
}