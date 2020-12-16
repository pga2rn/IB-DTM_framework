package core
//
//import (
//	"github.com/pga2rn/ib-dtm_framework/simulator/config"
//	"github.com/pga2rn/ib-dtm_framework/simulator/sim-map"
//	"github.com/pga2rn/ib-dtm_framework/simulator/vehicle"
//)
//
//////// life cycle //////
//// activate a vehicle into the map
//func ActivateVehicles (id uint64) error {
//	// if the active vehicles exceed the maximum, return
//	if m.MapStatus.ActiveVehiclesNum > m.SimConfig.VehicleNumMax{
//		return nil
//	}
//
//	// search for valid vehicle ID
//	index := -1
//	for i, v := range m.Vehicles{
//		if v == nil {
//			index = i
//			break
//		} else if v.vehicleStatus != Active {
//			index = i
//			break
//		}
//	}
//
//	// all the slot is activated (which is very rare)
//	if index < 0 {
//		return error()
//	}
//
//	// construct the vehicle
//	// pos
//	v.Pos.X = config.R.Intn(int(m.SimConfig.XLen))
//	v.Pos.Y = config.R.Intn(int(m.SimConfig.YLen))
//	// status
//	v.vehicleStatus = Active
//	v.lastMovementDirection = DirectionArray[config.R.Intn(len(DirectionArray))]
//
//	// register
//	v.id = uint64(index) // the same as the slot in the map
//	v.m = m // pointer to the map
//	m.Vehicles[index] = v // register into the map
//}
//// vehicle destroy
//// not really destroy, but just set it as Inactive
//func (v *Vehicle) Inactivate() error {
//	if v.vehicleStatus != InActive {
//		v.vehicleStatus = InActive
//		// reset the vehicles position
//		v.Pos = simmap.Position{X:-1, Y:-1}
//		v.lastMovementDirection = InActive
//		v.m = nil
//	} else {
//		return error()
//	}
//	return nil
//}
//
//////// simulation //////
//// exceed boundary test is executed by the caller function
//func (v *Vehicle) moveHelper(direction int){
//	// update pos
//	switch direction{
//	case XForward:
//		v.Pos.X += 1
//	case XBackward:
//		v.Pos.X -= 1
//	case YForward:
//		v.Pos.Y += 1
//	case YBackward:
//		v.Pos.Y -= 1
//	}
//	// update the vehicle's status accordingly
//	v.lastMovementDirection = direction
//}
//
//
//func (v *Vehicle) Move() error {
//	// if the vehicle is not activated
//	if v.vehicleStatus != Active {
//		return error()
//	}
//
//	// TODO: make the movement more scientifically in the future
//	for {
//		direction := DirectionArray[config.R.Intn(len(DirectionArray))]
//		switch direction {
//		case -v.lastMovementDirection:
//			// It is strange to move backward immediately
//			continue
//		case XForward:
//			if v.Pos.X+1 < int(v.m.SimConfig.XLen) {
//				v.moveHelper(direction)
//			} else {
//				// The vehicle drives out of the map
//				v.Inactivate()
//			}
//			return nil
//		case XBackward:
//			if v.Pos.X-1 > 0 {
//				v.moveHelper(direction)
//			} else {
//				v.Inactivate()
//			}
//			return nil
//		case YForward:
//			if v.Pos.Y+1 < int(v.m.SimConfig.YLen) {
//				v.moveHelper(direction)
//			} else {
//				v.Inactivate()
//			}
//			return nil
//		case YBackward:
//			if v.Pos.Y-1 > 0 {
//				v.moveHelper(direction)
//			} else {
//				v.Inactivate()
//			}
//			return nil
//		}
//	}
//
//	// after the movement, the vehicle will either
//	// 1. remain active, means it moves
//	// 2. being inactive, means it moves out of the map
//}