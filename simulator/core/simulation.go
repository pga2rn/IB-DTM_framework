package core

import (
	"context"
	"errors"
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
	"time"
)

func (sim *SimulationSession) SlotDeadline(slot uint64) time.Time {
	return timeutil.NextSlotTime(sim.Config.Genesis, slot)
}

func (sim *SimulationSession) Done() {
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

// contact and init with external RSU module
// DEPRECATED!: useless if we decouple RSU with validator
func (sim *SimulationSession) InitExternalRSUModule() error {
	// TODO: implement external RSU module contact
	return nil
}

// initializing vehicles, place VehicleNumMin vehicles into the network
func (sim *SimulationSession) WaitForVehiclesInit() error {
	logutil.LoggerList["core"].Debugf("[WaitForVehiclesInit] ..")
	if ok := sim.InitVehicles(); !ok {
		err := errors.New("failed to init vehicles")
		return err
	}
	return nil
}

////// simulation routines //////
// process slot
// routine:
// 1. move vehicles!
// 2. generate trust value offset!
func (sim *SimulationSession) ProcessSlot(ctx context.Context, slot uint64) error {
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
		// execute dtm logic
		sim.executeDTMLogic(SlotCtx, slot)

		// debug:
		count := 0
		for i := 0; i < sim.Config.VehicleNumMax; i++ {
			if sim.ActiveVehiclesBitMap.Get(i) {
				count++
			}
		}

		logutil.LoggerList["core"].
			Debugf("[ProcessSlot] active vehicles: %v, bitmap count %v",
				sim.ActiveVehiclesNum,
				count,
			)
		cancel()
		return nil
	}
}

// process epoch
// routine:
// 1. reassign the compromised RSU
// 2. reassign the misbehavior vehicles
// 3. gather previous epoch's data
// 2. calculated trust value and stored in the session

// TODO: mutex should be applied to trustvalue storage
func (sim *SimulationSession) ProcessEpoch(ctx context.Context, slot uint64) error {
	logutil.LoggerList["core"].Debugf("[ProcessEpoch] entering ..")
	select {
	case <-ctx.Done():
		logutil.LoggerList["core"].Debugf("[ProcessEpoch] context canceled")
		return errors.New("context canceled")
	default:
		// reassign the compromised RSU
		sim.CompromisedRSUBitMap = bitmap.NewTS(sim.Config.RSUNum)
		sim.initAssignCompromisedRSU(ctx)

		// reassign the misbehavior vehicle
		sim.MisbehaviorVehicleBitMap = bitmap.NewTS(sim.Config.VehicleNumMax)
		sim.InitAssignMisbehaveVehicle(ctx)

		// calculate trust value

		// debug
		logutil.LoggerList["core"].
			Debugf("[ProcessEpoch] epoch: %v, mdvp: %v, crsup: %v",
				sim.Epoch,
				sim.MisbehaviorVehiclePortion,
				sim.CompromisedRSUPortion,
			)

	}

	return nil
}
