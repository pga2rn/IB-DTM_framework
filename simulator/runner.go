package core

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
	"time"
)

// init the simulation session
func Run(ctx context.Context) *SimulationSession {
	cfg := config.GenYangNetConfig()
	cfg.SetGenesis(time.Now().Add(3 * time.Second))
	session := PrepareSimulationSession(cfg)

	go session.run(ctx)
	return session
}

func Done(session *SimulationSession) {
	session.done()
}

func (sim *SimulationSession) done() {
	// terminate the ticker
	sim.Ticker.Done()
	return
}

// start the simulation!
// routines are as follow:
// if checkpoint: processepoch
// processslot
// gather reports
func (sim *SimulationSession) run(ctx context.Context) {
	// init vehicles
	if err := sim.WaitForVehiclesInit(); err != nil {
		sim.done()
		logutil.LoggerList["core"].Fatal("Could not init vehicles: %v", err)
	}

	// init RSU
	// wait for every RSU to comes online
	if err := sim.WaitForRSUInit(); err != nil {
		sim.done()
		logutil.LoggerList["core"].Fatal("External RSU module is not ready: %v", err)
	}

	// wait for the blockchain
	// WaitForBlockchainStart
	// Ignored it! I will manually start blockchain and simulator

	// wait for statistics collecting modules

	// start the main loop
	logutil.LoggerList["core"].Debugf("[Run] Genesis kicks start!")
	for {
		ctx, cancel := context.WithCancel(ctx)

		select {
		case <-ctx.Done():
			logutil.LoggerList["core"].Debugf("Context canceled, stop the simulation.")
			cancel()
			return
		// the ticker will tick a uint64 slot index very slot
		case slot := <-sim.Ticker.C():
			logutil.LoggerList["core"].Debugf("[SlotTicker] Slot %v", slot)

			// check if the session's epoch and slot record is correct
			if slot != timeutil.SlotsSinceGenesis(sim.Config.Genesis) {
				// we are slower than the ticker, skipped some slots
				logutil.LoggerList["core"].
					Debugf("[Run] we are asynced with the ticker, %v, %v", slot, timeutil.SlotsSinceGenesis(sim.Config.Genesis))
				// catch up with the ticker
				logutil.LoggerList["core"].Debugf("[Run] catch up with the ticker")
			}
			// update the slot and epoch tracing in session before hand
			sim.Slot = timeutil.SlotsSinceGenesis(sim.Config.Genesis)
			sim.Epoch = timeutil.EpochsSinceGenesis(sim.Config.Genesis)

			// the following process must be finished within the slot
			slotCtx, cancel := context.WithDeadline(ctx, sim.SlotDeadline(slot))
			if err := sim.ProcessSlot(slotCtx, slot); err != nil {
				cancel()
				logutil.LoggerList["core"].Fatalf("failed to process slot: %v", err)
			}

			// if it is the checkpoint, or the start point of epoch
			if slot%sim.Config.SlotsPerEpoch == 0 {
				if err := sim.ProcessEpoch(slotCtx, slot); err != nil {
					cancel()
					logutil.LoggerList["core"].Fatalf("failed to process epoch: %v", err)
				}
				// spawn a new go routine for gathering reports for epoch
				go func() {
					// the report generation should be done before the next epoch
					epochCtx, cancel :=
						context.WithDeadline(ctx, timeutil.NextEpochTime(sim.Config.Genesis, slot))
					if err := sim.PrepareReportPerEpoch(epochCtx, slot); err != nil {
						logutil.LoggerList["core"].Debugf("failed to gather reports for epoch: %v", err)
					}
					cancel()
				}()
			}

			// spawn a new go routine to collect per slot report
			go func() {
				slotCtx, cancel := context.WithDeadline(ctx, sim.SlotDeadline(slot))
				if err := sim.PrepareReportPerSlot(slotCtx, slot); err != nil {
					logutil.LoggerList["core"].Debugf("failed to gather reports for slot: %v", err)
				}
				cancel()
			}()

			cancel() // terminate ctx for this slot
		} // slot
	} // main loop
}
