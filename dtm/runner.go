package dtm

import (
	"context"
	"errors"
	"github.com/pga2rn/ib-dtm_framework/shared"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
)

func Run(ctx context.Context) {
	session := DTMLogicSession{}
	session.run(ctx)
}

func (session *DTMLogicSession) done() {
	close(session.ChanBlockchain)
	close(session.ChanSim)
}

func (session *DTMLogicSession) WaitForSimulator(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return errors.New("[WaitForSimulator] context canceled")
	case v := <-session.ChanSim:
		// unpack
		pack := v.(shared.SimInitDTMCommunication)

		session.Vehicles = pack.Vehicles
		session.RSUs = pack.RSUs
		session.MisbehavingVehicleBitMap = pack.MisbehavingVehicleBitMap
		session.vmu, session.rmu = pack.Vmu, pack.Rmu
		return nil
	}
}

func (session *DTMLogicSession) run(ctx context.Context) {
	// wait for simulator to call & initialized the dtm logic module
	deadline, cancel := context.WithDeadline(ctx, session.SimConfig.Genesis)
	if err := session.WaitForSimulator(deadline); err != nil {
		cancel()
		session.done()
		logutil.LoggerList["dtm"].Fatalf("failed to wait for simulator start")
	}
	cancel()

	// after initialization is finished, waiting for the communication from the simulator
	for {
		select {
		case <-ctx.Done():
			session.done()
			return
		case v := <-session.ChanSim:
			// unpack
			pack := v.(shared.SimDTMCommunication)
			session.Slot, session.Epoch = pack.Slot, pack.Slot/session.SimConfig.SlotsPerEpoch
			session.CompromisedRSUBitMap = pack.CompromisedRSUBitMap

			logutil.LoggerList["dtm"].Debugf("[run] slot %v, epoch %v", session.Slot, session.Epoch)
			// init context
			// must be finished within a slot, otherwise the storage of RSU will be altered in the new epoch
			slotCtx, cancel :=
				context.WithDeadline(ctx, timeutil.SlotDeadline(session.SimConfig.Genesis, session.Slot))

			// init storage
			session.initDataStructureForEpoch(session.Epoch)
			// execute dtm logic
			session.genTrustValue(slotCtx, session.Slot)
			session.flagMisbehavingVehicle(slotCtx, session.Slot)
			// cancel the context for this epoch's process
			cancel()
			logutil.LoggerList["dtm"].Debugf("[run] slot %v, epoch %v done", session.Slot, session.Epoch)

			// emit a signal to tell the simulator to go on
			session.ChanSim <- true

			// sending raw results' pointer to the statistics module
			// including the raw trust value list and misbehaving vehicle bitmap
		}
	}
}
