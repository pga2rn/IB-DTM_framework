package simulator

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

// wait for rsu data structure ready
func (sim *SimulationSession) WaitForRSUInit() error {
	logutil.LoggerList["core"].Debugf("[WaitForRSUInit] ..")
	if ok := sim.InitRSUs(); !ok {
		return errors.New("failed to init RSU")
	}
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
	logutil.LoggerList["core"].Debugf("[ProcessSlot] slot %v", slot)
	SlotCtx, cancel := context.WithCancel(ctx)

	select {
	case <-ctx.Done():
		logutil.LoggerList["core"].Debugf("[ProcessSlot] context canceled.")
		cancel()
		return errors.New("context canceled")
	default:
		// move the vehicles!
		sim.moveVehiclesPerSlot(SlotCtx)
		// generate trust value offsets
		sim.executeDTMLogicPerSlot(SlotCtx, slot)

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
	logutil.LoggerList["core"].Debugf("[ProcessEpoch] processing epoch %v", slot/sim.Config.SlotsPerEpoch)
	select {
	case <-ctx.Done():
		logutil.LoggerList["core"].Debugf("[ProcessEpoch] context canceled")
		return errors.New("context canceled")
	default:
		// only init misbehaving vehicles at the start
		// TODO: init logic may be placed to other place
		if slot == 0 {
			sim.MisbehaviorVehicleBitMap = bitmap.NewTS(sim.Config.VehicleNumMax)
			sim.InitAssignMisbehaveVehicle(ctx)
		}

		// reassign the compromised RSU
		sim.CompromisedRSUBitMap = bitmap.NewTS(sim.Config.RSUNum)
		sim.initAssignCompromisedRSU(ctx)

		// calculate trust value
		//sim.genTrustValue(ctx, slot)
		//

		// debug
		logutil.LoggerList["core"].
			Debugf("[ProcessEpoch] mdvp: %v, crsup: %v",
				sim.MisbehaviorVehiclePortion,
				sim.CompromisedRSUPortion,
			)
		logutil.LoggerList["core"].
			Debugf("[ProcessEpoch] active vehicles %v", sim.ActiveVehiclesNum)

	}

	return nil
}
