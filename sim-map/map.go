package simmap

import (
	"github.com/pga2rn/ib-dtm_framework/config"
	"sync"
)

// each cross represents a CROSS within the map,
// which holds a RSU and 0 or more vehicles
// map looks like this:
// 0->N
// |
// v
// M
type cross struct {
	// a list of vehicle that appears
	Vehicles *sync.Map // map[uint64]*vehicle.Vehicle
}

type Map struct {
	// a 2d array represents the map
	Cross [][]*cross
}

func (c *cross) initCross() {
	c.Vehicles = &sync.Map{} // map[uint64]*vehicle.Vehicle)
}

// create a brand new map
func CreateMap(cfg *config.Config) *Map {
	m := &Map{}

	// prepare the map
	m.Cross = make([][]*cross, cfg.YLen)
	for i := range m.Cross {
		m.Cross[i] = make([]*cross, cfg.XLen)
		// init cross
		for j := 0; j < int(cfg.XLen); j++ {
			c := cross{}
			c.initCross()
			m.Cross[i][j] = &c
		}
	}

	return m
}
