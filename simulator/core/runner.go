package core

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
)

// start the simulation!
// routines are as follow:
// case 1: main process exit, simulation stop
// case 2: waiting for next slot
//		r1: update the map, move the vehicles
//		r2: generate trust value for newly moved vehicles
//		r3:	call RSU, provide trust value offsets to them and let them do the job
//		r2: calculate trust value
func (sim *SimulationSession) Run(ctx context.Context) {
	cleanup := sim.Done
	defer cleanup()

	// init vehicles
	if err := sim.WaitForVehiclesInit(); err != nil {
		cleanup()
		logutil.LoggerList["core"].Fatal("Could not init vehicles: %v", err)
	}

	// init RSU
	// wait for every RSU to comes online
	if err := sim.WaitForRSUInit(); err != nil {
		cleanup()
		logutil.LoggerList["core"].Fatal("External RSU module is not ready: %v", err)
	}

	// wait for the blockchain
	// WaitForBlockchainStart
	// Ignored it! I will manually start blockchain and simulator

	// process the genesis epoch
	// must be processed until genesis
	epochCtx, cancel := context.WithDeadline(ctx, sim.Config.Genesis)
	defer cancel()
	sim.ProcessEpoch(epochCtx, 0)

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

			// the following process must be finished within the slot
			deadline := sim.SlotDeadline(slot)
			slotCtx, cancel := context.WithDeadline(ctx, deadline)
			defer cancel()
			if err := sim.ProcessSlot(slotCtx, slot); err != nil {
				logutil.LoggerList["core"].Debugf("failed to process slot")
				return
			}

		//	// move the vehicles
		//	if err := sim.MoveVehicles(slotCtx, slot); err != nil {
		//		log.Fatal("Failed to move vehicles: %v", err)
		//		cancel()
		//	}
		//
		//	// generate the trust value offsets
		//	if err := sim.GenerateTrustValueOffsets(slotCtx, slot); err != nil {
		//		log.Fatal("Failed to generate trust value offsets: %v", err)
		//		break
		//	}
		//
		//	// if it is the checkpoint
		//	if slot % sim.Config.SlotsPerEpoch == 0 {
		//		if err := sim.ProcessEpoch(slotCtx, slot); err != nil {
		//			log.Fatal("Failed to process epoch: %v", err)
		//			break
		//		}
		//	}
		//
		//	// after the above function is completed, update the slot index
		//	sim.Slot = slot
		//
		//	// dispatch the trust value offsets to every RSU
		//	// call RSU to execute: 1. internal logics, 2. external logics
		//	if err := sim.ExecuteRSULogic(); err != nil {
		//		log.Fatal("Failed to execute RSU logics: %v", err)
		//	}
		}
	}

}
