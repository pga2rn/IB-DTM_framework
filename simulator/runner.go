package simulator

import (
	"context"
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

// start the simulation!
func (sim *SimulationSession) Run(ctx context.Context) {
	genesisCtx, cancel := context.WithDeadline(ctx, sim.Config.Genesis)

	// init vehicles
	if err := sim.WaitForVehiclesInit(genesisCtx); err != nil {
		sim.Done(ctx)
		logutil.GetLogger(PackageName).Fatal("could not init vehicles: %v", err)
	}

	// init RSU
	// wait for every RSU to comes online
	if err := sim.WaitForRSUInit(genesisCtx); err != nil {
		sim.Done(ctx)
		logutil.GetLogger(PackageName).Fatal("external RSU module is not ready: %v", err)
	}
	cancel()

	// start the main loop
	logutil.GetLogger(PackageName).Debugf("[Run] genesis kicks start!")
	for {
		select {
		case <-ctx.Done():
			logutil.GetLogger(PackageName).Debugf("context canceled, stop the simulation.")
			return
		// the ticker will tick a uint32 slot index very slot
		case slot := <-sim.Ticker.C():
			logutil.GetLogger(PackageName).Debugf("[simulator] Slot %v", slot)

			// check if the session's epoch and slot record is correct
			if slot != timeutil.SlotsSinceGenesis(sim.Config.Genesis) {
				logutil.GetLogger(PackageName).
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
