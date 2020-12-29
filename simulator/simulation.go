package simulator

import (
	"context"
	"errors"
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
)

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
func (sim *SimulationSession) ProcessEpoch(ctx context.Context, slot uint64) error {
	logutil.LoggerList["core"].Debugf("[ProcessEpoch] processing epoch %v", slot/sim.Config.SlotsPerEpoch)
	select {
	case <-ctx.Done():
		logutil.LoggerList["core"].Debugf("[ProcessEpoch] context canceled")
		return errors.New("context canceled")
	default:
		switch slot {
		case uint64(0):
			sim.MisbehaviorVehicleBitMap = bitmap.NewTS(sim.Config.VehicleNumMax)
			sim.InitAssignMisbehaveVehicle(ctx)
			// TODO: call the dtm module for init
		default:
			// TODO: call the dtm module for executing the previous epoch before clean up
			// reassign the compromised RSU
			sim.CompromisedRSUBitMap = bitmap.NewTS(sim.Config.RSUNum)
			sim.initAssignCompromisedRSU(ctx)
		}

		// wait for dtm logic module finish its job
		for {
			if v, ok := <-sim.ChanDTM; v.(bool) && ok {
				logutil.LoggerList["core"].Debugf("[processEpoch] dtm logic finished")
				break
			}
		}

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
