package simulator

import (
	"context"
	"errors"
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/shared"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
)

// simulator logics of simulator

// wait for rsu data structure ready
func (sim *SimulationSession) WaitForRSUInit() error {
	logutil.LoggerList["simulator"].Debugf("[WaitForRSUInit] ..")
	if ok := sim.InitRSUs(); !ok {
		return errors.New("failed to init RSU")
	}
	return nil
}

// initializing vehicles, place VehicleNumMin vehicles into the network
func (sim *SimulationSession) WaitForVehiclesInit() error {
	logutil.LoggerList["simulator"].Debugf("[WaitForVehiclesInit] ..")
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
func (sim *SimulationSession) ProcessSlot(ctx context.Context, slot uint32) error {
	logutil.LoggerList["simulator"].Debugf("[ProcessSlot] slot %v", slot)
	SlotCtx, cancel := context.WithCancel(ctx)

	select {
	case <-ctx.Done():
		logutil.LoggerList["simulator"].Debugf("[ProcessSlot] context canceled.")
		cancel()
		return errors.New("context canceled")
	default:
		// move the vehicles!
		sim.moveVehiclesPerSlot(SlotCtx, slot)
		// generate trust value offsets
		sim.executeDTMLogicPerSlot(SlotCtx, slot)

		cancel()
		return nil
	}

}

func (sim *SimulationSession) dialDTMLogicModulePerEpoch(ctx context.Context, slot uint32) {
	logutil.LoggerList["simulator"].Debugf("[dialDTMLogicModulePerEpoch] epoch %v", slot/sim.Config.SlotsPerEpoch-1)
	select {
	case <-ctx.Done():
		return
	default:
		pack := shared.SimDTMEpochCommunication{}
		pack.Slot, pack.CompromisedRSUBitMap = slot, sim.CompromisedRSUBitMap
		sim.ChanDTM <- pack
		// wait for dtm logic module to finish
		<-sim.ChanDTM
		logutil.LoggerList["simulator"].Debugf("[dialDTMLogicModulePerEpoch] dtm logic module finished")
	}

}

// process epoch
func (sim *SimulationSession) ProcessEpoch(ctx context.Context, slot uint32) error {
	epoch := slot / sim.Config.SlotsPerEpoch
	if epoch != 0 {
		epoch -= 1
	}
	logutil.LoggerList["simulator"].Debugf("[ProcessEpoch] processing epoch %v", epoch)

	select {
	case <-ctx.Done():
		logutil.LoggerList["simulator"].Debugf("[ProcessEpoch] context canceled")
		return errors.New("context canceled")
	default:
		switch slot {
		case uint32(0):
			// both misbehaving vehicles and compromised RSU will be assigned only at the beginning of the simulation
			sim.MisbehaviorVehicleBitMap = bitmap.NewTS(sim.Config.VehicleNumMax)
			sim.InitAssignMisbehaveVehicle(ctx)

			sim.CompromisedRSUBitMap = bitmap.NewTS(sim.Config.RSUNum)
			sim.initAssignCompromisedRSU(ctx)

			// signal the dtm logic module to init
			if err := sim.WaitForDTMLogicModule(); err != nil {
				sim.done()
				logutil.LoggerList["simulator"].Fatal("dtm logic module failed to init", err)
			}
		default:
			// call the dtm module for executing the previous epoch before new cRSU assignment
			sim.dialDTMLogicModulePerEpoch(ctx, slot)

			// debug
			logutil.LoggerList["simulator"].
				Infof("[ProcessEpoch] mdvp: %v, crsup: %v",
					sim.MisbehaviorVehiclePortion,
					sim.CompromisedRSUPortion,
				)
			logutil.LoggerList["simulator"].
				Infof("[ProcessEpoch] active vehicles %v", sim.ActiveVehiclesNum)
		}
	}

	return nil
}
