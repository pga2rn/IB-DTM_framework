package core

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/vehicle"
)

////// simulation //////
// a helper function to sync the vehicle status between session and vehicle object
func (sim *SimulationSession) UpdateVehicleStatus(v *vehicle.Vehicle, pos vehicle.Position, status int) {
	//logutil.LoggerList["core"].Debugf("vs %v, bitms %v, status %v",
	//	v.VehicleStatus, sim.ActiveVehiclesBitMap.Get(int(v.Id)), status)
	switch {
	case v.VehicleStatus == vehicle.InActive && status == vehicle.Active:
		// REMEMBER TO UPDATE THE VEHICLE'S STATUS!
		v.VehicleStatus = status
		// update the session
		sim.ActiveVehiclesNum += 1
		sim.ActiveVehiclesBitMap.Set(int(v.Id), true)
		// add the vehicle into the map
		sim.Map.Cross[pos.X][pos.Y].Vehicles.Store(v.Id, v)
	case v.VehicleStatus == vehicle.Active && status == vehicle.InActive:
		// REMEMBER TO UPDATE THE VEHICLE'S STATUS! AGAIN!
		v.VehicleStatus = status
		// update the session
		sim.ActiveVehiclesNum -= 1
		sim.ActiveVehiclesBitMap.Set(int(v.Id), false)
		// unregister the vehicle from the map
		sim.Map.Cross[pos.X][pos.Y].Vehicles.Delete(v.Id)
		// reset the vehicle after remove it from the map
		v.ResetVehicle()
	}
}

func (sim *SimulationSession) InitVehicles() bool {
	// activate the very first VehiclesNumMin vehicles
	sim.ActiveVehiclesNum = sim.Config.VehicleNumMin

	// init activated vehicles
	for i := 0; i < int(sim.Config.VehicleNumMin); i++ {
		v := vehicle.InitVehicle(
			uint64(i),
			sim.Config.XLen, sim.Config.YLen,
			vehicle.Active,
			sim.R,
		)

		// register the vehicle to the session
		sim.Vehicles[i] = v
		sim.ActiveVehiclesBitMap.Set(i, true)

		//logutil.LoggerList["core"].Debugf("pos %v", v.Pos)
		// place the vehicle onto the map
		sim.Map.Cross[v.Pos.X][v.Pos.Y].Vehicles.Store(uint64(i), v)
	}

	// init inactivate vehicles
	for i := sim.Config.VehicleNumMin; i < sim.Config.VehicleNumMax; i++ {
		v := vehicle.InitVehicle(
			uint64(i),
			sim.Config.XLen, sim.Config.YLen,
			vehicle.InActive,
			sim.R,
		)

		// register the vehicle to the session
		sim.Vehicles[i] = v
		sim.ActiveVehiclesBitMap.Set(i, false)
	}

	// init all vehicles' trust value
	//sim.AccurateTrustValueList = make([]float32, sim.Config.VehicleNumMax)
	//for i := range sim.AccurateTrustValueList {
	//	sim.AccurateTrustValueList[i] = 0
	//}
	//sim.BiasedTrustValueList = make([]float32, sim.Config.VehicleNumMax)
	//for i := range sim.BiasedTrustValueList {
	//	sim.BiasedTrustValueList[i] = 0
	//}

	return true
}

func (sim *SimulationSession) moveVehiclesPerSlot(ctx context.Context) {
	logutil.LoggerList["core"].Debugf("[moveVehiclesPerSlot] entering..")
	select {
	case <-ctx.Done():
		logutil.LoggerList["core"].Debugf("[moveVehiclesPerSlot] context canceled")
		return
	default:
		// activate extra vehicles
		interval := sim.Config.VehicleNumMax - sim.ActiveVehiclesNum
		newTarget := sim.R.RandIntRange(interval/3, interval*2/3)
		c := make(chan bool)
		// emit signal in the background to activate vehicles
		go func() {
			for i := 0; i < newTarget; i++ {
				c <- true
			}
			close(c)
		}()

		// randomly pick vehicle, iterating the whole list
		for _, i := range sim.R.Perm(int(sim.Config.VehicleNumMax)) {
			go func(i int) {
				v := sim.Vehicles[i]
				switch sim.ActiveVehiclesBitMap.Get(i) {
				case true:
					sim.moveVehicle(v)
				case false:
					for { // activate new vehicles
						if f, ok := <-c; ok && f {
							// when we activate a new vehicle
							// first we update the vehicle object
							v.EnterMap(sim.R, sim.Config.XLen, sim.Config.YLen)
							// then we update it onto the map
							sim.UpdateVehicleStatus(v, v.Pos, vehicle.Active)
							// finally we move it!
							sim.moveVehicle(v)
						}
						return
					}
				}
			}(i) // go routine
		}
	}
}

// mark a specific vehicle as inactive, and unregister it from the map
func (sim *SimulationSession) inactivateVehicle(v *vehicle.Vehicle, oldPos vehicle.Position) {
	// filter out invalid vehicle
	if v == nil || v.Id > uint64(sim.Config.VehicleNumMax) {
		return
	}
	sim.UpdateVehicleStatus(v, oldPos, vehicle.InActive)
}

// move vehicle from one cross to another
func (sim *SimulationSession) updateVehiclePos(v *vehicle.Vehicle, oldPos vehicle.Position) {
	// unregister the vehicle from the old cross
	sim.Map.Cross[oldPos.X][oldPos.Y].Vehicles.Delete(v.Id)
	// register the vehicle into the new cross
	sim.Map.Cross[v.Pos.X][v.Pos.Y].Vehicles.Store(v.Id, v)
}

// move a single vehicle
// routine:
// 1. move the vehicle!
// 2. update vehicle's position record within the map
func (sim *SimulationSession) moveVehicle(v *vehicle.Vehicle) {
	// if the vehicle is not activated
	if v.VehicleStatus != vehicle.Active {
		return
	}

	// update the vehicle object
	newDirection := v.MovementDecisionMaker(sim.R, sim.Config.XLen, sim.Config.YLen)
	oldPos := v.Pos
	v.MoveHelper(newDirection)

	// after the vehicle object is updated,
	// check whether the vehicle is still in the map
	switch {
	case v.Pos.X >= 0 && v.Pos.Y >= 0 && v.Pos.X < sim.Config.XLen && v.Pos.Y < sim.Config.YLen:
		// finally update the vehicle's status on map
		sim.updateVehiclePos(v, oldPos)
	default:
		// the vehicle moves out of the map
		sim.inactivateVehicle(v, oldPos)
	}

}

func (sim *SimulationSession) InitAssignMisbehaveVehicle(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	default:
		count := 0
		sim.MisbehaviorVehiclePortion = sim.R.RandFloatRange(
			sim.Config.MisbehaveVehiclePortionMin,
			sim.Config.MisbehaveVehiclePortionMax,
		)
		// assign roles to vehicles no matter what status it is
		target := int(float32(sim.Config.VehicleNumMax) * sim.MisbehaviorVehiclePortion)

		for count < target {
			index := sim.R.RandIntRange(0, sim.Config.VehicleNumMax)
			if !sim.MisbehaviorVehicleBitMap.Get(index) {
				sim.MisbehaviorVehicleBitMap.Set(index, true)
				count++
			}
		}
	}
}
