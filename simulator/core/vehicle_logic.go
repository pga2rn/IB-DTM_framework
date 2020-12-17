package core

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/randutil"
	"github.com/pga2rn/ib-dtm_framework/simulator/vehicle"
)

////// simulation //////
// try to put the vehicle close to the center of the map
func (sim *SimulationSession) genNewPosition() vehicle.Position{
	xl, yl := int(sim.Config.XLen), int(sim.Config.YLen)
	return vehicle.Position{
		X: randutil.RandIntRange(sim.R, xl / 8, xl * 7 / 8),
		Y: randutil.RandIntRange(sim.R, yl / 8, xl * 7 / 8),
	}
}

// a helper function to sync the vehicle status between session and vehicle object
func (sim *SimulationSession) UpdateVehicleStatus(v *vehicle.Vehicle, status int){
	count := 0
	for i := 0; i < sim.Config.VehicleNumMax; i++ {
		if sim.ActiveVehiclesBitMap.Get(i){
			count ++
		}
	}
	logutil.LoggerList["core"].Debugf("[UpdateVehicleStatus] id %v, v.status %v, status %v, bitmap status %v, bitmap count %v, av %v",
		v.Id,
		v.VehicleStatus,
		status,
		sim.ActiveVehiclesBitMap.Get(int(v.Id)),
		count,
		sim.ActiveVehiclesNum,
		)

	switch {
	case v.VehicleStatus == vehicle.InActive && status == vehicle.Active:
		sim.ActiveVehiclesNum += 1
		sim.ActiveVehiclesBitMap.Set(int(v.Id), true)
	case v.VehicleStatus == vehicle.Active && status == vehicle.InActive:
		// counter alter
		sim.ActiveVehiclesNum -= 1
		sim.ActiveVehiclesBitMap.Set(int(v.Id), false)
		// unregister the vehicle from the map
		delete(sim.Map.Cross[v.Pos.X][v.Pos.Y].Vehicles, v.Id)
	}
	// reset the vehicle pos and lastmovement
	v.ResetVehicle()
}

func (sim *SimulationSession) InitVehicles() bool {
	// activate the very first VehiclesNumMin vehicles
	sim.ActiveVehiclesNum = sim.Config.VehicleNumMin

	// init activated vehicles
	for i := 0 ; i < int(sim.Config.VehicleNumMin); i++ {
		v := &vehicle.Vehicle{}
		v.InitVehicle(
			uint64(i),
			sim.genNewPosition(),
			vehicle.Active,
			vehicle.NotMove,
		)

		// register the vehicle to the session
		sim.Vehicles[i] = v
		sim.ActiveVehiclesBitMap.Set(i, true)

		//logutil.LoggerList["core"].Debugf("pos %v", v.Pos)
		// place the vehicle onto the map
		sim.Map.Cross[v.Pos.X][v.Pos.Y].Vehicles[uint64(i)] = v
	}

	// init inactivate vehicles
	for i := sim.Config.VehicleNumMin; i < sim.Config.VehicleNumMax; i++ {
		v := &vehicle.Vehicle{}
		v.InitVehicle(
			uint64(i),
			vehicle.Position{},
			vehicle.InActive,
			vehicle.NotMove,
		)

		// register the vehicle to the session
		sim.Vehicles[i] = v
		sim.ActiveVehiclesBitMap.Set(i, false)
	}

	// init all vehicles' trust value
	sim.AccurateTrustValueList = make([]float32, sim.Config.VehicleNumMax)
	for i := range sim.AccurateTrustValueList {
		sim.AccurateTrustValueList[i] = 0
	}
	sim.BiasedTrustValueList = make([]float32, sim.Config.VehicleNumMax)
	for i := range sim.BiasedTrustValueList {
		sim.BiasedTrustValueList[i] = 0
	}

	return true
}

func (sim *SimulationSession) moveVehicles(ctx context.Context) {
	logutil.LoggerList["core"].Debugf("[moveVehicles] entering..")
	select {
	case <-ctx.Done():
		logutil.LoggerList["core"].Debugf("[moveVehicles] context canceled")
		return
	default:
		// activate extra vehicles
		interval := sim.Config.VehicleNumMax - sim.ActiveVehiclesNum
		newCount, newTarget := 0, randutil.RandIntRange(sim.R, interval / 3, interval * 2 / 3)

		// randomly pick vehicle, iterating the whole list
		for _, i := range sim.R.Perm(int(sim.Config.VehicleNumMax)) {
			v := sim.Vehicles[i]
			switch sim.ActiveVehiclesBitMap.Get(i) {
			case true:
					sim.moveVehicle(v)
			case false:
					if newCount < newTarget {
						sim.UpdateVehicleStatus(v, vehicle.Active)
						newCount ++

						v.Pos, v.LastMovementDirection = sim.genNewPosition(), vehicle.NotMove
						sim.moveVehicle(v)
					}
			}
		}
	}
}

// mark a specific vehicle as inactive, and unregister it from the map
func (sim *SimulationSession) inactivateVehicle(v *vehicle.Vehicle) {
	// filter out invalid vehicle
	if v == nil || v.Id > uint64(sim.Config.VehicleNumMax) {
		return
	}
	sim.UpdateVehicleStatus(v, vehicle.InActive)
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
		case -v.LastMovementDirection, vehicle.NotMove:
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

func (sim *SimulationSession) InitAssignMisbehaveVehicle(ctx context.Context){
	select {
	case <-ctx.Done():
		return
	default:
		count := 0
		sim.MisbehaviorVehiclePortion = randutil.RandFloatRange(
			sim.R,
			sim.Config.MisbehaveVehiclePortionMin,
			sim.Config.MisbehaveVehiclePortionMax,
		)
		// assign roles to vehicles no matter what status it is
		target := int(float32(sim.Config.VehicleNumMax) * sim.MisbehaviorVehiclePortion)

		for count < target {
			index := randutil.RandIntRange(sim.R, 0, sim.Config.VehicleNumMax)
			if ! sim.MisbehaviorVehicleBitMap.Get(index) {
				sim.MisbehaviorVehicleBitMap.Set(index, true)
				count ++
			}
		}
	}
}