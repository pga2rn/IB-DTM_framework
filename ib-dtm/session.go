package ib_dtm

import (
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/rsu"
	"github.com/pga2rn/ib-dtm_framework/shared/fwtype"
	"github.com/pga2rn/ib-dtm_framework/shared/randutil"
	"github.com/pga2rn/ib-dtm_framework/shared/timeutil"
	"sync"
)

type IBDTMSession struct {
	IBDTMConfig   *config.IBDTMConfig
	SimConfig     *config.SimConfig
	ExpConfigList []*config.ExperimentConfig

	// chain for every proposal experiments
	Blockchain map[string]*BlockchainRoot

	// pointer to the vehicles and RSU
	RSUs                 [][]*rsu.RSU
	CompromisedRSUBitMap *bitmap.Threadsafe
	rmu                  *sync.Mutex

	// inter-modules communication
	ChanSim, ChanDTM chan interface{}

	Ticker      timeutil.Ticker
	Epoch, Slot uint32

	// beacon status
	BeaconStatus map[string]*BeaconStatus

	// latest trust value results
	TrustValueStorage map[string]*fwtype.TrustValuesPerEpoch

	// a random source
	R *randutil.RandUtil
}

// should only include proposal experiments
func PrepareBlockchainModule(
	simCfg *config.SimConfig, expCfgList []*config.ExperimentConfig, ibdtmCfg *config.IBDTMConfig,
	sim, dtm chan interface{}) *IBDTMSession {

	session := &IBDTMSession{
		SimConfig:     simCfg,
		ExpConfigList: expCfgList,
		IBDTMConfig:   ibdtmCfg,
		ChanDTM:       dtm,
		ChanSim:       sim,
	}

	// init data structure
	session.Blockchain = make(map[string]*BlockchainRoot)
	session.BeaconStatus = make(map[string]*BeaconStatus)

	// init storage area for each experiment
	// temporary storage for latest trust value
	session.TrustValueStorage = make(map[string]*fwtype.TrustValuesPerEpoch)
	for _, exp := range session.ExpConfigList {
		session.Blockchain[exp.Name] = InitBlockchain()

		// init beaconstatus for each experiment
		session.BeaconStatus[exp.Name] = InitBeaconStatus(
			simCfg, ibdtmCfg, exp, session.Blockchain[exp.Name])
	}

	// prepare the ticker
	session.Ticker = timeutil.GetSlotTicker(simCfg.Genesis, simCfg.SecondsPerSlot)

	// prepare the random source
	session.R = randutil.InitRand(123)

	return session
}
