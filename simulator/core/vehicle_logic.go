package core

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/simulator/vehicle"
)

////// simulation //////
func (sim *SimulationSession) moveVehicles(ctx context.Context) {
	logutil.LoggerList["core"].Debugf("[moveVehicles] entering..")
	select {
	case <-ctx.Done():
		logutil.LoggerList["core"].Debugf("[moveVehicles] context canceled")
		return
	default:
		// check if num of vehicles is insufficient
		for _, v := range sim.Vehicles {
			if v.VehicleStatus != vehicle.Active {
				if sim.ActiveVehiclesNum < sim.Config.VehicleNumMin {
					v.VehicleStatus = vehicle.Active
					v.ResetVehicle() // reset the pos and lastmovement
					sim.ActiveVehiclesNum += 1
				}
			}
			if v.VehicleStatus == vehicle.Active {
				sim.moveVehicle(v)
			}
		}
	}

}

// mark a specific vehicle as inactive, and unregister it from the map
func (sim *SimulationSession) inactivateVehicle(v *vehicle.Vehicle) {
	// filter out invalid vehicle
	if v == nil && v.Id > sim.Config.VehicleNumMax {
		return
	}
	v.VehicleStatus = vehicle.InActive
	sim.ActiveVehiclesBitMap.Set(int(v.Id), false)
	// unregister the vehicle from the old cross
	delete(sim.Map.Cross[v.Pos.X][v.Pos.Y].Vehicles, v.Id)
}

// move vehicle from one cross to another
func (sim *SimulationSession) updateVehiclePos(v *vehicle.Vehicle) {
	// unregister the vehicle from the old cross
	delete(sim.Map.Cross[v.Pos.X][v.Pos.Y].Vehicles, v.Id)
	// register the vehicle into the new cross
	sim.Map.Cross[v.Pos.X][v.Pos.Y].Vehicles[v.Id] = v
}

// move a single vehicle
// routine:
// 1. move the vehicle!
// 2. update vehicle's position record within the map
// TODO: optimize the following code?
func (sim *SimulationSession) moveVehicle(v *vehicle.Vehicle) {
	// if the vehicle is not activated
	if v.VehicleStatus != vehicle.Active {
		return
	}

	// TODO: make the movement more scientifically in the future
	for {
		direction := vehicle.DirectionArray[sim.R.Intn(len(vehicle.DirectionArray))]
		switch direction {
		case -v.LastMovementDirection:
			// It is strange to move backward immediately
			// or stay still all the time
			continue
		case vehicle.XForward:
			if v.Pos.X+1 < int(sim.Config.XLen) {
				v.MoveHelper(direction)
				sim.updateVehiclePos(v)
			} else {
				// The vehicle drives out of the map
				sim.inactivateVehicle(v)
			}
			return
		case vehicle.XBackward:
			if v.Pos.X-1 > 0 {
				v.MoveHelper(direction)
				sim.updateVehiclePos(v)
			} else {
				sim.inactivateVehicle(v)
			}
			return
		case vehicle.YForward:
			if v.Pos.Y+1 < int(sim.Config.YLen) {
				v.MoveHelper(direction)
				sim.updateVehiclePos(v)
			} else {
				sim.inactivateVehicle(v)
			}
			return
		case vehicle.YBackward:
			if v.Pos.Y-1 > 0 {
				v.MoveHelper(direction)
				sim.updateVehiclePos(v)
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
