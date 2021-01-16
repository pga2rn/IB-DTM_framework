package simmap

import (
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/shared/fwtype"
	"github.com/pga2rn/ib-dtm_framework/vehicle"
	"sync"
)

// each cross represents a CROSS within the map,
// which holds a RSU and 0 or more vehiclesMap
type cross struct {
	// a list of vehicle that appears
	mu           sync.RWMutex
	vehiclesMap  map[uint32]*vehicle.Vehicle // map[uint32]*vehicle.Vehicle
	vehiclesList bitmap.Bitmap
}

// a 2d array represents the map
type SimMap [][]*cross

func (m SimMap) GetCross(pos fwtype.Position) *cross {
	return m[pos.X][pos.Y]
}

func createCross(vnum int) *cross {
	return &cross{
		mu:           sync.RWMutex{},
		vehiclesMap:  make(map[uint32]*vehicle.Vehicle), // map[uint32]*vehicle.Vehicle)
		vehiclesList: bitmap.New(vnum),
	}
}

func (c *cross) RemoveVehicle(vid uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.vehiclesMap, vid)
	c.vehiclesList.Set(int(vid), false)
}

func (c *cross) AddVehicle(vid uint32, v *vehicle.Vehicle) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.vehiclesMap[vid] = v
	c.vehiclesList.Set(int(vid), true)
}

// get list of vehiclesMap at current slot
func (c *cross) GetVehicleList() *[]uint32 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	res := make([]uint32, len(c.vehiclesMap))
	for i := range c.vehiclesMap {
		res = append(res, i)
	}
	return &res
}

func (c *cross) GetVehicleNum() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.vehiclesMap)
}

func (c *cross) CheckIfVehicleInManagementZone(vid uint32) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.vehiclesList.Get(int(vid))
}

// create a brand new map
func CreateMap(cfg *config.SimConfig) SimMap {
	m := make(SimMap, cfg.YLen)

	// prepare the map
	for i := range m {
		m[i] = make([]*cross, cfg.XLen)
		// init cross
		for j := 0; j < cfg.XLen; j++ {
			m[i][j] = createCross(cfg.VehicleNumMax)
		}
	}

	return m
}
