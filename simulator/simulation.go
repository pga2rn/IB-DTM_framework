package simulator

import (
	"context"
	"errors"
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
)

// simulator logics of simulator

// wait for rsu data structure ready
func (sim *SimulationSession) WaitForRSUInit(ctx context.Context) error {
	logutil.GetLogger(PackageName).Debugf("[WaitForRSUInit] ..")
	select {
	case <-ctx.Done():
		return errors.New("context canceled")
	default:
		if ok := sim.InitRSUs(); !ok {
			return errors.New("failed to init RSU")
		}
	}
	return nil
}

// initializing vehicles, place VehicleNumMin vehicles into the network
func (sim *SimulationSession) WaitForVehiclesInit(ctx context.Context) error {
	logutil.GetLogger(PackageName).Debugf("[WaitForVehiclesInit] ..")
	select {
	case <-ctx.Done():
		return errors.New("context canceled")
	default:
		if ok := sim.InitVehicles(); !ok {
			return errors.New("failed to init vehicles")
		}
	}
	return nil
}

////// simulation routines //////
// process slot
// routine:
// 1. move vehicles!
// 2. generate trust value offset!
func (sim *SimulationSession) ProcessSlot(ctx context.Context, slot uint32) {
	logutil.GetLogger(PackageName).Debugf("[ProcessSlot] slot %v", slot)
	SlotCtx, cancel := context.WithCancel(ctx)

	select {
	case <-ctx.Done():
		logutil.GetLogger(PackageName).Fatalf("[ProcessSlot] context canceled.")
	default:
		// move the vehicles!
		sim.moveVehiclesPerSlot(SlotCtx, slot)

		// generate trust value offsets
		sim.prepareRSUsForSlot(SlotCtx, slot)
		// generate and dispatch trust value offsets to every RSUs
		sim.genTrustValueOffset(SlotCtx, slot)
		// execute related RSU logic
		sim.forgeTrustValueOffsets(SlotCtx, slot)

		// signal the ib-dtm module to execute related logic
		sim.dialIBDTMLogicModulePerSlot(SlotCtx, slot)
	}
	cancel()
}

// process epoch
func (sim *SimulationSession) ProcessEpoch(ctx context.Context, slot uint32) {
	epoch := slot / sim.Config.SlotsPerEpoch
	if epoch != 0 {
		epoch -= 1
	}
	logutil.GetLogger(PackageName).Debugf("[ProcessEpoch] processing epoch %v", epoch)

	select {
	case <-ctx.Done():
		logutil.GetLogger(PackageName).Fatalf("[ProcessEpoch] context canceled")
	default:
		switch slot {
		case uint32(0):
			// both misbehaving vehicles and compromised RSU will be assigned only at the beginning of the simulation
			sim.MisbehaviorVehicleBitMap = bitmap.NewTS(sim.Config.VehicleNumMax)
			sim.initAssignMisbehaveVehicle(ctx)

			sim.CompromisedRSUBitMap = bitmap.NewTS(sim.Config.RSUNum)
			sim.initAssignCompromisedRSU(ctx)

			// at the start of the simulation, init other module
			// dial and init the ib-dtm module
			sim.dialInitIBDTMModule(ctx)
			// signal the dtm logic module to init
			sim.dialInitDTMLogicModule()

			// debug
			logutil.GetLogger(PackageName).
				Infof("[ProcessEpoch] mdvp: %v, crsup: %v",
					sim.MisbehaviorVehiclePortion,
					sim.CompromisedRSUPortion,
				)
		default:
			// call the dtm module for executing the previous epoch before new cRSU assignment
			sim.dialDTMLogicModulePerEpoch(ctx, slot)

			// debug
			logutil.GetLogger(PackageName).
				Infof("[ProcessEpoch] done! active vehicles %v", sim.ActiveVehiclesNum)
		}
	}
}
