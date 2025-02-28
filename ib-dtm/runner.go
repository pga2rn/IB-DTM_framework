package ib_dtm

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
)

func (session *IBDTMSession) Done(ctx context.Context) {
	session.Ticker.Done()
}

func (session *IBDTMSession) Run(ctx context.Context) {
	logutil.GetLogger(PackageName).Debugf("[Run] start!")

	// wait for simulator to activate the dtm logic module
	if err := session.WaitForSimulator(ctx); err != nil {
		session.Done(ctx)
		logutil.GetLogger(PackageName).Fatalf("failed to wait for simulator start")
	}

	logutil.GetLogger(PackageName).Debugf("[Run] genesis kics start!")
	for {
		select {
		case <-ctx.Done():
			logutil.GetLogger(PackageName).Fatal("[Run] context canceled, abort")
		case slot := <-session.Ticker.C():
			logutil.GetLogger(PackageName).Debugf("[blockchain] slot %v", slot)

			// prepare blockchain head block
			for _, bc := range session.Blockchain {
				if _, err := bc.InitBlockchainBlock(slot, session.IBDTMConfig); err != nil {
					logutil.GetLogger(PackageName).Fatal("[Run] failed to init new block, slot %v", slot)
				}
			}

			// sync
			session.Epoch, session.Slot = slot/session.SimConfig.SlotsPerEpoch, slot

			// process the logic
			slotCtx, cancel :=
				context.WithDeadline(ctx, timeutil.SlotDeadline(session.SimConfig.Genesis, slot))

			// if the slot is the checkpoint slot
			if slot%session.SimConfig.SlotsPerEpoch == 0 {
				session.processEpoch(slotCtx, slot)
			}

			session.ProcessSlot(slotCtx, slot)

			cancel()
		}
	}
}
