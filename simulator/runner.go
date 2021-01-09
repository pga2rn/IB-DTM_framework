package simulator

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
)

// main entry of simulator
func (sim *SimulationSession) Done(ctx context.Context) {
	// terminate the ticker
	sim.Ticker.Done()
	close(sim.ChanDTM)
	return
}

func (sim *SimulationSession) WaitForDTMLogicModule() {
	initPack := shared.SimInitDTMCommunication{
		MisbehavingVehicleBitMap: sim.MisbehaviorVehicleBitMap,
		RSUs:                     &sim.RSUs,
		Rmu:                      &sim.rmu,
	}

	// send the init pack to the dtm logic module
	logutil.LoggerList["simulator"].Debugf("[WaitForDTMLogicModule] send init pack to dtm logic module")
	sim.ChanDTM <- initPack

	// wait for the dtm logic module finishing the init
	<-sim.ChanDTM
	logutil.LoggerList["simulator"].Debugf("[WaitForDTMLogicModule] dtm logic module init finished")
}

// start the simulation!
func (sim *SimulationSession) Run(ctx context.Context) {
	genesisCtx, cancel := context.WithDeadline(ctx, sim.Config.Genesis)

	// init vehicles
	if err := sim.WaitForVehiclesInit(genesisCtx); err != nil {
		sim.Done(ctx)
		logutil.LoggerList["simulator"].Fatal("could not init vehicles: %v", err)
	}

	// init RSU
	// wait for every RSU to comes online
	if err := sim.WaitForRSUInit(genesisCtx); err != nil {
		sim.Done(ctx)
		logutil.LoggerList["simulator"].Fatal("external RSU module is not ready: %v", err)
	}
	cancel()

	// start the main loop
	logutil.LoggerList["simulator"].Debugf("[Run] genesis kicks start!")
	for {
		select {
		case <-ctx.Done():
			logutil.LoggerList["simulator"].Debugf("context canceled, stop the simulation.")
			return
		// the ticker will tick a uint32 slot index very slot
		case slot := <-sim.Ticker.C():
			logutil.LoggerList["simulator"].Debugf("[simulator] Slot %v", slot)

			// check if the session's epoch and slot record is correct
			if slot != timeutil.SlotsSinceGenesis(sim.Config.Genesis) {
				logutil.LoggerList["simulator"].
					Fatalf("[Run] we are asynced with the ticker, %v, %v", slot, timeutil.SlotsSinceGenesis(sim.Config.Genesis))
			}
			sim.Slot = timeutil.SlotsSinceGenesis(sim.Config.Genesis)
			sim.Epoch = timeutil.EpochsSinceGenesis(sim.Config.Genesis)

			slotCtx, cancel := context.WithDeadline(ctx, timeutil.SlotDeadline(sim.Config.Genesis, slot))
			// if it is the checkpoint, or the start point of epoch
			if slot%sim.Config.SlotsPerEpoch == 0 {
				sim.ProcessEpoch(slotCtx, slot)
			}

			sim.ProcessSlot(slotCtx, slot)

			cancel() // terminate ctx for this slot
		} // slot
	} // main loop
}
