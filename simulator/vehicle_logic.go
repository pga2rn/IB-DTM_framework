package simulator

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/pair"
	"github.com/pga2rn/ib-dtm_framework/vehicle"
	"sync"
)

////// simulation //////
// a helper function to sync the vehicle status between session and vehicle object
func (sim *SimulationSession) UpdateVehicleStatus(v *vehicle.Vehicle, pos pair.Position, status int) {
	switch {
	case v.VehicleStatus == vehicle.InActive && status == vehicle.Active:
		// REMEMBER TO UPDATE THE VEHICLE'S STATUS!
		v.VehicleStatus = status
		// update the session
		sim.ActiveVehiclesNum += 1
		sim.ActiveVehiclesBitMap.Set(int(v.Id), true)
		// add the vehicle into the map
		sim.Map.GetCross(pos).Vehicles.Store(v.Id, v)
	case v.VehicleStatus == vehicle.Active && status == vehicle.InActive:
		// REMEMBER TO UPDATE THE VEHICLE'S STATUS! AGAIN!
		v.VehicleStatus = status
		// update the session
		sim.ActiveVehiclesNum -= 1
		sim.ActiveVehiclesBitMap.Set(int(v.Id), false)
		// unregister the vehicle from the map
		sim.Map.GetCross(pos).Vehicles.Delete(v.Id)
		// reset the vehicle after remove it from the map
		v.ResetVehicle()
	}
}

func (sim *SimulationSession) InitVehicles() bool {
	// activate the very first VehiclesNumMin vehicles
	sim.ActiveVehiclesNum = sim.Config.VehicleNumMin

	// init activated vehicles
	for i := 0; i < sim.Config.VehicleNumMin; i++ {
		v := vehicle.InitVehicle(
			uint32(i),
			sim.Config.XLen, sim.Config.YLen,
			vehicle.Active,
			sim.R,
		)

		// register the vehicle to the session
		sim.Vehicles[i] = v
		sim.ActiveVehiclesBitMap.Set(i, true)

		// place the vehicle onto the map
		sim.Map.GetCross(v.Pos).AddVehicle(uint32(i), v)
	}

	// init inactivate vehicles
	for i := sim.Config.VehicleNumMin; i < sim.Config.VehicleNumMax; i++ {
		v := vehicle.InitVehicle(
			uint32(i),
			sim.Config.XLen, sim.Config.YLen,
			vehicle.InActive,
			sim.R,
		)

		// register the vehicle to the session
		sim.Vehicles[i] = v
		sim.ActiveVehiclesBitMap.Set(i, false)
	}
	return true
}

func (sim *SimulationSession) moveVehiclesPerSlot(ctx context.Context, slot uint32) {
	select {
	case <-ctx.Done():
		logutil.LoggerList["simulator"].Fatalf("[moveVehiclesPerSlot] context canceled")
	default:
		// sync
		wg := sync.WaitGroup{}

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
		for _, i := range sim.R.Perm(sim.Config.VehicleNumMax) {
			wg.Add(1)

			go func(i int) {
				select {
				case <-ctx.Done():
					wg.Done()
					logutil.LoggerList["simulator"].Debugf("[moveVehiclesPerSlot] go routine context canceled detected")
					return
				default:
					v := sim.Vehicles[i]
					switch sim.ActiveVehiclesBitMap.Get(i) {
					case true:
						sim.moveVehicle(v)
					case false:
						for {
							if val, ok := <-c; ok && val { // activate new vehicles
								// when we activate a new vehicle
								// first we update the vehicle object
								v.EnterMap(sim.R, sim.Config.XLen, sim.Config.YLen)
								// then we update it onto the map
								sim.UpdateVehicleStatus(v, v.Pos, vehicle.Active)
								// finally we move it!
								sim.moveVehicle(v)
							} else { // no more new vehicles are needed
								break
							}
						}
					}
				} // select
				wg.Done()
			}(i) // go routine
		} // for loop

		// wait for all jobs to be done
		wg.Wait()
	} // select
}

// mark a specific vehicle as inactive, and unregister it from the map
func (sim *SimulationSession) inactivateVehicle(v *vehicle.Vehicle, oldPos pair.Position) {
	// filter out invalid vehicle
	if v == nil || v.Id > uint32(sim.Config.VehicleNumMax) {
		return
	}
	sim.UpdateVehicleStatus(v, oldPos, vehicle.InActive)
}

// move vehicle from one cross to another
// wrap the operation
func (sim *SimulationSession) updateVehiclePos(v *vehicle.Vehicle, oldPos pair.Position) {
	// unregister the vehicle from the old cross
	sim.Map.GetCross(oldPos).RemoveVehicle(v.Id)
	// register the vehicle into the new cross
	sim.Map.GetCross(v.Pos).AddVehicle(v.Id, v)
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
	v.VehicleMove(newDirection)

	// after the vehicle object is updated,
	// check whether the vehicle is still in the map
	switch {
	case v.Pos.X >= 0 && v.Pos.Y >= 0 && v.Pos.X < sim.Config.XLen && v.Pos.Y < sim.Config.YLen:
		// if the vehicle still in the map, update its position
		sim.updateVehiclePos(v, oldPos)
	default:
		// else inactivate the vehicle because it is out of the map
		sim.inactivateVehicle(v, oldPos)
	}

}

func (sim *SimulationSession) initAssignMisbehaveVehicle(ctx context.Context) {
	select {
	case <-ctx.Done():
		logutil.LoggerList["simulator"].Fatalf("[initAssignMisbehaveVehicle] context canceled")
		return
	default:
		sim.MisbehaviorVehiclePortion = sim.R.RandFloatRange(
			sim.Config.MisbehaveVehiclePortionMin,
			sim.Config.MisbehaveVehiclePortionMax,
		)
		// assign roles to vehicles no matter what status it is
		target := int(float32(sim.Config.VehicleNumMax) * sim.MisbehaviorVehiclePortion)

		for i := 0; i < target; i++ {
			index := sim.R.RandIntRange(0, sim.Config.VehicleNumMax)
			if !sim.MisbehaviorVehicleBitMap.Get(index) {
				sim.MisbehaviorVehicleBitMap.Set(index, true)
			}
		}
	}
}
