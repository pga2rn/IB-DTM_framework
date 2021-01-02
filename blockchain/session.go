package blockchain

import (
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
)

type BlockchainSession struct {
	Config *config.SimConfig

	Blockchain *BlockchainHead

	Ticker      timeutil.Ticker
	Epoch, Slot uint32
}

func PrepareBlockchainModule(cfg *config.SimConfig, sim, dtm chan interface{}) *BlockchainSession {
	bc := &BlockchainSession{}
	bc.Config = cfg

	// init storage area
	bc.Blockchain = InitBlockchain()

	// prepare the ticker
	bc.Ticker = timeutil.GetSlotTicker(cfg.Genesis, cfg.SecondsPerSlot)

	return bc
}
