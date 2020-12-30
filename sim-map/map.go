package simmap

import (
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/vehicle"
	"sync"
)

// each cross represents a CROSS within the map,
// which holds a RSU and 0 or more vehicles
type cross struct {
	// a list of vehicle that appears
	Vehicles     *sync.Map // map[uint64]*vehicle.Vehicle
	VehiclesList *bitmap.Threadsafe
}

type Map struct {
	// a 2d array represents the map
	Cross [][]*cross
}

func (c *cross) initCross(vnum int) {
	c.Vehicles = &sync.Map{} // map[uint64]*vehicle.Vehicle)
	c.VehiclesList = bitmap.NewTS(vnum)
}

func (c *cross) RemoveVehicle(vid uint64) {
	c.Vehicles.Delete(vid)
	c.VehiclesList.Set(int(vid), false)
}

func (c *cross) AddVehicle(vid uint64, v *vehicle.Vehicle) {
	c.Vehicles.Store(vid, v)
	c.VehiclesList.Set(int(vid), true)
}

func (c *cross) GetVehicleList() *[]uint64 {
	res := make([]uint64, 16, 32)
	for i := 0; i < c.VehiclesList.Len(); i++ {
		if c.VehiclesList.Get(i) {
			res = append(res, uint64(i))
		}
	}
	return &res
}

func (c *cross) GetVehicleNum() int {
	count := 0
	for i := 0; i < c.VehiclesList.Len(); i++ {
		if c.VehiclesList.Get(i) {
			count += 1
		}
	}
	return count
}

func (c *cross) CheckIfVehicleInManagementZone(vid uint64) bool {
	return c.VehiclesList.Get(int(vid))
}

// create a brand new map
func CreateMap(cfg *config.SimConfig) *Map {
	m := &Map{}

	// prepare the map
	m.Cross = make([][]*cross, cfg.YLen)
	for i := range m.Cross {
		m.Cross[i] = make([]*cross, cfg.XLen)
		// init cross
		for j := 0; j < int(cfg.XLen); j++ {
			c := cross{}
			c.initCross(cfg.VehicleNumMax)
			m.Cross[i][j] = &c
		}
	}

	return m
}
