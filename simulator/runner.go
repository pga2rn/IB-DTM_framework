package simulator

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
)

// main entry of simulator

func Done(session *SimulationSession) {
	session.done()
}

func (sim *SimulationSession) done() {
	// terminate the ticker
	sim.Ticker.Done()
	close(sim.ChanDTM)
	return
}

func (sim *SimulationSession) WaitForDTMLogicModule() error {
	initPack := shared.SimInitDTMCommunication{}
	initPack.MisbehavingVehicleBitMap = sim.MisbehaviorVehicleBitMap

	initPack.RSUs, initPack.Vehicles = &sim.RSUs, &sim.Vehicles
	initPack.Rmu, initPack.Vmu = &sim.rmu, &sim.vmu

	// send the init pack to the dtm logic module
	logutil.LoggerList["simulator"].Debugf("[WaitForDTMLogicModule] send init pack to dtm logic module")
	sim.ChanDTM <- initPack
	// wait for the dtm logic module finishing the init
	<-sim.ChanDTM
	logutil.LoggerList["simulator"].Debugf("[WaitForDTMLogicModule] dtm logic module init finished")
	return nil
}

// start the simulation!
func (sim *SimulationSession) Run(ctx context.Context) {
	// init vehicles
	if err := sim.WaitForVehiclesInit(); err != nil {
		sim.done()
		logutil.LoggerList["simulator"].Fatal("Could not init vehicles: %v", err)
	}

	// init RSU
	// wait for every RSU to comes online
	if err := sim.WaitForRSUInit(); err != nil {
		sim.done()
		logutil.LoggerList["simulator"].Fatal("External RSU module is not ready: %v", err)
	}

	// start the main loop
	logutil.LoggerList["simulator"].Debugf("[Run] Genesis kicks start!")
	for {
		ctx, cancel := context.WithCancel(ctx)

		select {
		case <-ctx.Done():
			logutil.LoggerList["simulator"].Debugf("Context canceled, stop the simulation.")
			cancel()
			return
		// the ticker will tick a uint32 slot index very slot
		case slot := <-sim.Ticker.C():
			logutil.LoggerList["simulator"].Debugf("[SlotTicker] Slot %v", slot)

			// check if the session's epoch and slot record is correct
			if slot != timeutil.SlotsSinceGenesis(sim.Config.Genesis) {
				// we are slower than the ticker, skipped some slots
				logutil.LoggerList["simulator"].
					Debugf("[Run] we are asynced with the ticker, %v, %v", slot, timeutil.SlotsSinceGenesis(sim.Config.Genesis))
				// catch up with the ticker
				logutil.LoggerList["simulator"].Debugf("[Run] catch up with the ticker")
			}
			// update the slot and epoch tracing in session before hand
			sim.Slot = timeutil.SlotsSinceGenesis(sim.Config.Genesis)
			sim.Epoch = timeutil.EpochsSinceGenesis(sim.Config.Genesis)

			// the following process must be finished within the slot
			slotCtx, cancel := context.WithDeadline(ctx, timeutil.SlotDeadline(sim.Config.Genesis, slot))

			// if it is the checkpoint, or the start point of epoch
			if slot%sim.Config.SlotsPerEpoch == 0 {
				if err := sim.ProcessEpoch(slotCtx, slot); err != nil {
					cancel()
					logutil.LoggerList["simulator"].Fatalf("failed to process epoch: %v", err)
				}
			}

			if err := sim.ProcessSlot(slotCtx, slot); err != nil {
				cancel()
				logutil.LoggerList["simulator"].Fatalf("failed to process slot: %v", err)
			}

			cancel() // terminate ctx for this slot
		} // slot
	} // main loop
}
