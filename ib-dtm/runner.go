package ib_dtm

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
)

func (session *IBDTMSession) Done() {
	session.Ticker.Done()
}

func (session *IBDTMSession) Run(ctx context.Context) {
	logutil.LoggerList["ib-dtm"].Debugf("[Run] genesis kics start!")
	for {
		select {
		case <-ctx.Done():
			logutil.LoggerList["ib-dtm"].Fatal("[Run] context canceled, abort")
		case slot := <-session.Ticker.C():
			logutil.LoggerList["ib-dtm"].Debugf("[blockchain] slot %v", slot)

			// prepare blockchain head block
			for _, bc := range session.Blockchain {
				if _, err := bc.InitBlockchainBlock(slot, session.SimConfig); err != nil {
					logutil.LoggerList["ib-dtm"].Fatal("[Run] failed to init new block, slot %v", slot)
				}
			}

			// wait for the signal from sim
			if slot != <-session.ChanSim {
				logutil.LoggerList["ib-dtm"].Fatalf("[Run] async with simulator")
			}
			session.Epoch, session.Slot = slot/session.SimConfig.SlotsPerEpoch, session.Slot

			// process the logic
			slotCtx, cancel :=
				context.WithDeadline(ctx, timeutil.SlotDeadline(session.SimConfig.Genesis, slot))

			if slot/session.SimConfig.SlotsPerEpoch == 0 {
				// checkpoint slot is in the new epoch
				if slot == 0 {
					session.processEpoch(slotCtx, 0)
				} else {
					session.processEpoch(slotCtx, session.Epoch-1)
				}
			}

			session.ProcessSlot(slotCtx, slot)
			cancel()
		}
	}
}
