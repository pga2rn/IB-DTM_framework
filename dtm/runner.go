package dtm

import (
	"context"
	"errors"
	"github.com/pga2rn/ib-dtm_framework/shared"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
)

func (session *DTMLogicSession) Done(ctx context.Context) {
	close(session.ChanIBDTM)
	close(session.ChanSim)
}

func (session *DTMLogicSession) WaitForSimulator(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return errors.New("[WaitForSimulator] context canceled")
	case v := <-session.ChanSim:
		// unpack
		pack := v.(shared.SimInitDTMCommunication)

		session.RSUs = pack.RSUs
		session.MisbehavingVehicleBitMap = pack.MisbehavingVehicleBitMap
		session.rmu = pack.Rmu

		// after the init, signal the simulator
		session.ChanSim <- true
		logutil.LoggerList["dtm"].Debugf("[WaitForSimulator] init finished!")
		return nil
	}
}

func (session *DTMLogicSession) Run(ctx context.Context) {
	logutil.LoggerList["dtm"].Debugf("[Run] start!")

	// wait for simulator to activate the dtm logic module
	if err := session.WaitForSimulator(ctx); err != nil {
		session.Done(ctx)
		logutil.LoggerList["dtm"].Fatalf("failed to wait for simulator start")
	}

	// after initialization is finished, waiting for the communication from the simulator
	for {
		select {
		case <-ctx.Done():
			logutil.LoggerList["dtm"].Fatalf("[Run] context canceled")
		case v := <-session.ChanSim:
			// unpack
			pack := v.(shared.SimDTMEpochCommunication)
			session.Epoch = pack.Slot/session.SimConfig.SlotsPerEpoch - 1
			session.CompromisedRSUBitMap = pack.CompromisedRSUBitMap
			session.ActiveVehiclesNum = pack.ActiveVehiclesNum

			logutil.LoggerList["dtm"].Debugf("[dtm] epoch %v", session.Epoch)
			// init context
			// must be finished within a slot, otherwise the storage of RSU will be altered in the new epoch
			slotCtx, cancel :=
				context.WithDeadline(ctx, timeutil.SlotDeadline(session.SimConfig.Genesis, pack.Slot))

			// init storage
			session.initDataStructureForEpoch(session.Epoch)
			// execute dtm logic
			session.genBaselineTrustValue(slotCtx, session.Epoch)
			session.genProposalTrustValue(slotCtx, session.Epoch)

			session.flagMisbehavingVehicles(slotCtx, session.Epoch)
			session.genStatistics(slotCtx, session.Epoch)

			// inform the rpc server the newest results
			session.informRPCServer(slotCtx, session.Epoch)

			// cancel the context for this epoch's process
			cancel()
			logutil.LoggerList["dtm"].Debugf("[Run] epoch %v Done", session.Epoch)

			// emit a signal to tell the simulator to go on
			session.ChanSim <- true

			// sending raw results' pointer to the statistics module
			// including the raw trust value list and misbehaving vehicle bitmap
		}
	}
}
