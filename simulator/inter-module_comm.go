package simulator

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
)

func (sim *SimulationSession) dialInitDTMLogicModule() {
	initPack := &shared.SimInitDTMCommunication{
		MisbehavingVehicleBitMap: sim.MisbehaviorVehicleBitMap,
		RSUs:                     sim.RSUs,
		Rmu:                      &sim.rmu,
	}

	// send the init pack to the dtm logic module
	logutil.LoggerList["simulator"].Debugf("[dialInitDTMLogicModule] send init pack to dtm logic module")
	sim.ChanDTM <- initPack

	// wait for the dtm logic module finishing the init
	<-sim.ChanDTM
	logutil.LoggerList["simulator"].Debugf("[dialInitDTMLogicModule] dtm logic module init finished")
}

func (sim *SimulationSession) dialIBDTMLogicModulePerSlot(ctx context.Context, slot uint32) {
	logutil.LoggerList["simulator"].Debugf("[dialIBDTMLogicModulePerSlot] slot %v", slot)

	select {
	case <-ctx.Done():
		logutil.LoggerList["simulator"].Fatalf("[dialIBDTMLogicModulePerSlot] context canceled")
	default:
		// signal the ib-dtm module with slot
		sim.ChanIBDTM <- slot
		<-sim.ChanIBDTM
	}
}

func (sim *SimulationSession) dialInitIBDTMModule(ctx context.Context) {
	defer logutil.LoggerList["simulator"].Debugf("[dialInitIBDTMModule] done!")
	select {
	case <-ctx.Done():
		logutil.LoggerList["simulator"].Fatalf("[dialInitIBDTMModule] failed to init the IBDTM module")
	default:
		pack := shared.SimInitIBDTMCommunication{
			RSUs:                 sim.RSUs,
			Rmu:                  &sim.rmu,
			CompromisedRSUBitMap: sim.CompromisedRSUBitMap,
		}
		sim.ChanIBDTM <- &pack
		<-sim.ChanIBDTM
	}
}

func (sim *SimulationSession) dialDTMLogicModulePerEpoch(ctx context.Context, slot uint32) {
	logutil.LoggerList["simulator"].Debugf("[dialDTMLogicModulePerEpoch] epoch %v", slot/sim.Config.SlotsPerEpoch-1)
	select {
	case <-ctx.Done():
		logutil.LoggerList["simulator"].Fatalf("[dialDTMLogicModulePerEpoch] epoch %v, context canceled", slot/sim.Config.SlotsPerEpoch-1)
	default:
		pack := shared.SimDTMEpochCommunication{}
		pack.Slot, pack.ActiveVehiclesNum, pack.CompromisedRSUBitMap = slot, int32(sim.ActiveVehiclesNum), sim.CompromisedRSUBitMap
		sim.ChanDTM <- pack
		// wait for dtm logic module to finish
		<-sim.ChanDTM
		logutil.LoggerList["simulator"].Debugf("[dialDTMLogicModulePerEpoch] dtm logic module finished")
	}

}
