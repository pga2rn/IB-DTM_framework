package simmap

import (
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/shared/pair"
	"github.com/pga2rn/ib-dtm_framework/vehicle"
	"sync"
)

// each cross represents a CROSS within the map,
// which holds a RSU and 0 or more vehicles
type cross struct {
	// a list of vehicle that appears
	vehicles     *sync.Map // map[uint32]*vehicle.Vehicle
	vehiclesList *bitmap.Threadsafe
}

type Map struct {
	// a 2d array represents the map
	cross [][]*cross
}

func (m *Map) GetCross(pos pair.Position) *cross {
	return m.cross[pos.X][pos.Y]
}

func (c *cross) initCross(vnum int) {
	c.vehicles = &sync.Map{} // map[uint32]*vehicle.Vehicle)
	c.vehiclesList = bitmap.NewTS(vnum)
}

func (c *cross) RemoveVehicle(vid uint32) {
	c.vehicles.Delete(vid)
	c.vehiclesList.Set(int(vid), false)
}

func (c *cross) AddVehicle(vid uint32, v *vehicle.Vehicle) {
	c.vehicles.Store(vid, v)
	c.vehiclesList.Set(int(vid), true)
}

func (c *cross) GetVehicleList() *[]uint32 {
	res := make([]uint32, 16, 32)
	for i := 0; i < c.vehiclesList.Len(); i++ {
		if c.vehiclesList.Get(i) {
			res = append(res, uint32(i))
		}
	}
	return &res
}

func (c *cross) GetVehicleNum() int {
	count := 0
	for i := 0; i < c.vehiclesList.Len(); i++ {
		if c.vehiclesList.Get(i) {
			count += 1
		}
	}
	return count
}

func (c *cross) CheckIfVehicleInManagementZone(vid uint32) bool {
	return c.vehiclesList.Get(int(vid))
}

// create a brand new map
func CreateMap(cfg *config.SimConfig) *Map {
	m := &Map{}

	// prepare the map
	m.cross = make([][]*cross, cfg.YLen)
	for i := range m.cross {
		m.cross[i] = make([]*cross, cfg.XLen)
		// init cross
		for j := 0; j < int(cfg.XLen); j++ {
			c := cross{}
			c.initCross(cfg.VehicleNumMax)
			m.cross[i][j] = &c
		}
	}

	return m
}
