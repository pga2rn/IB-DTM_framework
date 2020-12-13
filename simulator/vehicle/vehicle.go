package vehicle

import (
	"github.com/pga2rn/ib-dtm_framework/simulator/sim-map"
)

// defined the data structure in a life cycle
type Vehicle struct {
	// unique id of each vehicle
	id uint64

	// current position of the vehicle
	// set to nil for inactive
	position simmap.Position
	// the path of vehicle movement, represented by pos
	// reset when becomes inactive
	path []simmap.Position

	// trust value related
	trust_value float32

	// vehicle status
	vehicle_status int
}

// vehiclestatus
const (
	Destroyed = -1
	NotInit   = -1 // the same as destroyed, the data structure is not available
	InActive  = 0  // temporary leave the map, but the data structure remains
	Active    = 1  // in the map right now
)


////// life cycle //////
// activate a vehicle into the map
func (v *Vehicle) Activate (m *simmap.Map) error {
	index := -1
	for i, v := range m.Vehicles{
		if v == nil {
			index = i
			break
		} else if v.vehicle_status != Active {
			index = i
			break
		}
	}

	if index < 0 {
		return error()
	}

	// register the vehicle to the map
	v.id = uint64(index) // the same as the slot in the map
	m.Vehicles[index] = v
}
// vehicle destroy
Done()

////// simulation //////
// simulate the movement within the map
Move()