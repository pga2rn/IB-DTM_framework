package core

import (
	"context"
	"errors"
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/shared/dtmutils"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/randutil"
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
	"github.com/pga2rn/ib-dtm_framework/simulator/vehicle"
	"time"
)

func (sim *SimulationSession) SlotDeadline(slot uint64) time.Time {
	return timeutil.NextSlotTime(sim.Config.Genesis, slot)
}

func (sim *SimulationSession) Done(){
	// terminate the ticker
	sim.Ticker.Done()
	return
}

// wait for rsu data structure ready
func (sim *SimulationSession) WaitForRSUInit() error {
	logutil.LoggerList["core"].Debugf("[WaitForRSUInit] ..")
	if ok := sim.InitRSU(); !ok {
		return errors.New("failed to init RSU")
	}
	if err := sim.InitExternalRSUModule(); err != nil {
		return errors.New("failed to finished external RSU module initializing")
	}
	return nil
}

func (sim *SimulationSession) InitRSU() bool {
	for x := range sim.RSUs {
		// init every RSU data structure
		for y := range sim.RSUs[x] {
			r := sim.RSUs[x][y]

			r.Id = uint64(y) * uint64(sim.Config.XLen) + uint64(x)
			r.Epoch = 0
			r.Slot = 0

			// start as un-compromised RSU
			r.CompromisedFlag = false
			// uploading tracker
			r.NextSlotForUpload = 0

			// init the data structure of trust value offset storage
			r.TrustValueOffsetPerSlot =
				make([]map[uint64]*dtmutils.TrustValueOffset, sim.Config.SlotsPerEpoch)
			for i := range r.TrustValueOffsetPerSlot{
				// init map structure for every slot
				r.TrustValueOffsetPerSlot[i] = make(map[uint64]*dtmutils.TrustValueOffset)
			}

			// not yet connected with external RSU module
			r.ExternalRSUModuleInitFlag = false
		}
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
	logutil.LoggerList["core"].Debugf("[WaitForVehiclesInit] ..")
	sim.ActiveVehiclesNum = sim.Config.VehicleNumMin
	if ok := sim.InitVehicles(); !ok {
		err := errors.New("failed to init vehicles")
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
	sim.AccurateTrustValueList = make([]float32, sim.Config.VehicleNumMax)
	for i := range sim.AccurateTrustValueList {
		sim.AccurateTrustValueList[i] = 1.0
	}
	sim.BiasedTrustValueList = make([]float32, sim.Config.VehicleNumMax)
	for i := range sim.BiasedTrustValueList {
		sim.BiasedTrustValueList[i] = 1.0
	}

	return true
}

////// simulation routines //////
// process slot
// routine:
// 1. move vehicles!
// 2. generate trust value offset!
func (sim *SimulationSession) ProcessSlot(ctx context.Context, slot uint64) error{
	logutil.LoggerList["core"].Debugf("[ProcessSlot] entering ..")
	SlotCtx, cancel := context.WithCancel(ctx)

	select {
	case <-ctx.Done():
		logutil.LoggerList["core"].Debugf("[ProcessSlot] context canceled.")
		cancel()
		return errors.New("context canceled")
	default:
		// move the vehicles!
		sim.moveVehicles(SlotCtx)
		// generate trust value offset for specific slot
		sim.genTrustValueOffset(SlotCtx, slot)
		return nil
	}
}

// process epoch
// routine:
// 1. reassign the compromised RSU
// 1. reassign the misbehavior vehicles
// 2. calculated trust value and stored in the session

// IMPORTANT: mutex should be applied to the reports storage
func (sim *SimulationSession) ProcessEpoch(ctx context.Context, slot uint64) error {
	// reassign the compromised RSU
	sim.CompromisedRSUBitMap = bitmap.New(sim.Config.RSUNum)

	count := 0
	sim.CompromisedRSUPortion = randutil.RandFloatRange(
			sim.R,
			sim.Config.PortionOfCompromisedRSUMin,
			sim.Config.PortionOfCompromisedRSUMax,
		)
	target := int(float32(sim.Config.RSUNum) * sim.CompromisedRSUPortion)

	for count < target {
		index := randutil.RandIntRange(sim.R, 0, sim.Config.RSUNum)
		if ! sim.CompromisedRSUBitMap.Get(index) {
			sim.CompromisedRSUBitMap.Set(index, true)
			sim.RSUs[index %][]
		}
	}




	//// compromised RSU count
	//compromisedRSUCount := 0
	//compromisedRSUTargetPortion := randutil.RandFloatRange(
	//	sim.R,
	//	sim.Config.PortionOfCompromisedRSUMin,
	//	sim.Config.PortionOfCompromisedRSUMax,
	//)
	//compromisedRSUTargetNum := int(float32(sim.Config.RSUNum) * compromisedRSUTargetPortion)
	//
	//for compromisedRSUCount < compromisedRSUTargetNum {
	//	x := randutil.RandIntRange(sim.R, 0, int(sim.Config.XLen))
	//	y := randutil.RandIntRange(sim.R, 0, int(sim.Config.YLen))
	//
	//	// turn a good RSU into a bad one
	//	r := sim.RSUs[x][y]
	//	switch r.CompromisedFlag{
	//	case true:
	//		continue
	//	case false:
	//		compromisedRSUCount += 1
	//		r.CompromisedFlag = true
	//	}
	//}
	//
	//
	//
	//
	//cRSUCount,  := 0,
	//for rsuCount {
	//
	//}

	return nil
}