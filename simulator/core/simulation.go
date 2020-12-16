package core

import (
	"context"
	"errors"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/simulator/dtm"
	"github.com/pga2rn/ib-dtm_framework/simulator/vehicle"
	"time"
)

func (sim *SimulationSession) SlotDeadline(slot uint64) time.Time {
	duration := time.Duration((slot + 1) * sim.Config.SecondsPerSlot) * time.Second
	return sim.Config.Genesis.Add(duration)
}

func (sim *SimulationSession) Done(){
	// terminate the ticker
	sim.Ticker.Done()
	return
}

// wait for rsu data structure ready
func (sim *SimulationSession) WaitForRSUInit() error {
	if ok := sim.InitRSU(); !ok {
		return errors.New("failed to init RSU")
	}
	if err := sim.InitExternalRSUModule(); err != nil {
		return errors.New("failed to finished external RSU module initializing")
	}
	return nil
}

func (sim *SimulationSession) InitRSU() bool {
	num := int(sim.Config.RSUNum)
	for i := 0; i < num; i++ {
		r := &dtm.RSU{}

		r.Id = uint64(i)
		r.Epoch = 0
		r.Slot = 0

		// start as un-compromised RSU
		r.CompromisedFlag = false
		// uploading tracker
		r.NextSlotForUpload = 0

		// not yet connected with external RSU module
		r.ExternalRSUModuleInitFlag = false

		// register the RSU
		sim.RSUs[i] = r
	}
	return true
}

// contact and init with external RSU module
// DEPRECATED!: useless if we decouple RSU with validator
func (sim *SimulationSession) InitExternalRSUModule() error {
	// TODO: implement external RSU module contact
	return nil
}

// initializing vehicles, place VehicleNumMin vehicles into the network
func (sim *SimulationSession) WaitForVehiclesInit() error {
	sim.ActiveVehiclesNum = sim.Config.VehicleNumMin
	if ok := sim.InitVehicles(); !ok {
		err := errors.New("Failed to init vehicles.")
		return err
	}
	return nil
}

func (sim *SimulationSession) InitVehicles() bool {
	// activate the very first VehiclesNumMin vehicles
	sim.ActiveVehiclesNum = sim.Config.VehicleNumMin

	// init activated vehicles
	for i := 0 ; i < int(sim.Config.VehicleNumMin); i++ {
		v := &vehicle.Vehicle{}
		v.Pos = vehicle.Position {
			sim.R.Intn(int(sim.Config.XLen)),
			sim.R.Intn(int(sim.Config.YLen)),
		}
		v.VehicleStatus = vehicle.Active
		v.LastMovementDirection = vehicle.NotMove

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
		v.VehicleStatus = vehicle.InActive
		v.LastMovementDirection = vehicle.NotMove

		// register the vehicle to the session
		sim.Vehicles[i] = v
		sim.ActiveVehiclesBitMap.Set(int(i), false)
	}

	// init all vehicles' trust value
	sim.AccurateTrustValueList = make([]float32, sim.Config.VehicleNumMax, 1.0)
	sim.BiasedTrustValueList = make([]float32, sim.Config.VehicleNumMax, 1.0)

	return true
}

////// simulation routines //////
// process slot
// routine:
// 1. move vehicles(including moving out of the map, set as inactive)
// 1.1 check if nums of vehicles less than vehiclenummin, or activate new vehicles
// 1.2 update rsu vehicles list
// 2. generate trust value offsets for every active vehicles
// 3. update rsu
// 3.1 update rsu trustvalueoffsetlist
func (sim *SimulationSession) ProcessSlot(ctx context.Context, slot uint64) error{
	SlotCtx, cancel := context.WithCancel(ctx)
	c := make(chan interface{})

	// move the vehicles!
	go sim.moveVehicles(SlotCtx, c)
	// update rsu trust value offset list
	go sim.executeRSU(SlotCtx, c)

	select {
	case <-ctx.Done():
		logutil.LoggerList["core"].Debugf("[ProcessSlot] context canceled.")
		cancel()
		return errors.New("context canceled")
	case <- c:
		// the 2 go routines finished
		return nil
	}
}



// process epoch
// routine:
// 1. process slot
// 2. calculated trust value and stored in the session
// 3. statistics gathering
func (sim *SimulationSession) ProcessEpoch(ctx context.Context, slot uint64){

}