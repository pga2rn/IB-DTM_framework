package blockchain

import (
	"context"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
)

func (session *BlockchainSession) Done() {
	session.Ticker.Done()
}

func (session *BlockchainSession) Run(ctx context.Context) {

	logutil.LoggerList["blockchain"].Debugf("[Run] genesis kics start!")
	for {
		select {
		case <-ctx.Done():
			logutil.LoggerList["blockchain"].Fatal("[Run] context canceled, abort")
		case slot := <-session.Ticker.C():
			logutil.LoggerList["blockchain"].Debugf("[blockchain] slot %v", slot)

			// TODO: keep implementing proposal logic
			if _, err := session.Blockchain.InitBlockchainBlock(slot, session.Config); err != nil {
				logutil.LoggerList["blockchain"].Fatal("[Run] failed to init new block")
			}
		}
	}
}
