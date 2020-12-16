package core

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/simulator/config"
	"github.com/pga2rn/ib-dtm_framework/simulator/vehicle"
)

////// simulation //////
func (sim *SimulationSession) moveVehicles(ctx context.Context,c chan interface{}) {
	// after the process is completed, call the caller to finish
	defer func(){
		c <- 1
	}()

	// check if num of vehicles is insufficient
	for _, v := range sim.Vehicles {
		if v.VehicleStatus != vehicle.Active {
			if sim.ActiveVehiclesNum < sim.Config.VehicleNumMin {
				v.VehicleStatus = vehicle.Active
				v.ResetVehicle() // reset the pos and lastmovement
			}
		}
		if v.VehicleStatus == vehicle.Active {
			sim.moveVehicle(v)
		}
	}

}

// mark a specific vehicle as inactive
func (sim *SimulationSession) inactivateVehicle(v *vehicle.Vehicle) {
	if v == nil && v.Id > sim.Config.VehicleNumMax {
		return
	}
	v.VehicleStatus = vehicle.InActive
	sim.ActiveVehiclesBitMap.Set(int(v.Id), false)
}

// move a single vehicle
func (sim *SimulationSession) moveVehicle(v *vehicle.Vehicle) {
	// if the vehicle is not activated
	if v.VehicleStatus != vehicle.Active {
		return
	}

	// TODO: make the movement more scientifically in the future
	for {
		direction := vehicle.DirectionArray[config.R.Intn(len(vehicle.DirectionArray))]
		switch direction {
		case -v.LastMovementDirection:
			// It is strange to move backward immediately
			// or stay still all the time
			continue
		case vehicle.XForward:
			if v.Pos.X+1 < int(sim.Config.XLen) {
				v.MoveHelper(direction)
			} else {
				// The vehicle drives out of the map
				sim.inactivateVehicle(v)
			}
			return
		case vehicle.XBackward:
			if v.Pos.X-1 > 0 {
				v.MoveHelper(direction)
			} else {
				sim.inactivateVehicle(v)
			}
			return
		case vehicle.YForward:
			if v.Pos.Y+1 < int(sim.Config.YLen) {
				v.MoveHelper(direction)
			} else {
				sim.inactivateVehicle(v)
			}
			return
		case vehicle.YBackward:
			if v.Pos.Y-1 > 0 {
				v.MoveHelper(direction)
			} else {
				sim.inactivateVehicle(v)
			}
			return
		}
	}

	// after the movement, the vehicle will either
	// 1. remain active, means it moves
	// 2. being inactive, means it moves out of the map
}
