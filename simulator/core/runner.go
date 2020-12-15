package core

import "context"

// start the simulation!
// routines are as follow:
// case 1: main process exit, simulation stop
// case 2: waiting for next slot
//		r1: update the map, move the vehicles
//		r2: generate trust value for newly moved vehicles
//		r3:	call RSU, provide trust value offsets to them and let them do the job
//		r2: calculate trust value
func run(ctx context.Context, sim *SimulationSession) {
	cleanup := sim.Done
	defer cleanup()

	// init RSU
	// wait for every RSU to comes online
	if err := sim.WaitForRSUInit(); err != nil {
		cleanup()
		log.Fatal("External RSU module is not ready: %v", err)
	}

	// init vehicles
	if err := sim.WaitForInitingVehicles(); err != nil {
		cleanup()
		log.Fatal("Could not init vehicles: %v", err)
	}

	// wait for the blockchain
	// WaitForBlockchainStart

	// start the main loop
	for {
		ctx, cancel := context.WithCancel(ctx)

		select {
		case <-ctx.Done():
			log.Info("Context canceled, stop the simulation.")
			cancel()
			return

		case slot := <-sim.Ticker.NextSlot():

			// the following process must be finished within the slot
			deadline := sim.SlotDeadline(slot)
			slotCtx, cancel := context.WithDeadline(ctx, deadline)
			log.WithField("deadline", deadline).
				Debug("The slot process must be finished within the current slot.")

			// move the vehicles
			if err := sim.MoveVehicles(slotCtx, slot); err != nil {
				log.Fatal("Failed to move vehicles: %v", err)
				cancel()
			}

			// generate the trust value offsets
			if err := sim.GenerateTrustValueOffsets(slotCtx, slot); err != nil {
				log.Fatal("Failed to generate trust value offsets: %v", err)
				break
			}

			// if it is the checkpoint
			if slot % sim.Config.SlotsPerEpoch == 0 {
				if err := sim.ProcessEpoch(slotCtx, slot); err != nil {
					log.Fatal("Failed to process epoch: %v", err)
					break
				}
			}

			// after the above function is completed, update the timestream
			sim.TimeStream = Beacon{
				Epoch: slot % sim.Config.SlotsPerEpoch,
				Slot:  slot,
			}

			// dispatch the trust value offsets to every RSU
			// call RSU to execute: 1. internal logics, 2. external logics
			if err := sim.ExecuteRSULogic(); err != nil {
				log.Fatal("Failed to execute RSU logics: %v", err)
			}
		}
	}

}
