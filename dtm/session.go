package dtm

import (
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/rsu"
	"github.com/pga2rn/ib-dtm_framework/shared/dtmtype"
	"github.com/pga2rn/ib-dtm_framework/shared/randutil"
	"github.com/pga2rn/ib-dtm_framework/vehicle"
	"sync"
)

// communicating with simulator: RSU compromised bitmap, slot

type DTMLogicSession struct {
	// configs
	Config    *map[string]*config.ExperimentConfig
	SimConfig *config.SimConfig
	// the correct answer
	MisbehavingVehicleBitMap *bitmap.Threadsafe

	// session status
	Slot, Epoch          uint64
	CompromisedRSUBitMap *bitmap.Threadsafe // only valid for specific epoch

	// pointer to the vehicles and RSU
	// I don't know if it is a good idea to use mutex via pointer
	Vehicles *[]*vehicle.Vehicle
	RSUs     *[][]*rsu.RSU
	vmu      *sync.Mutex
	rmu      *sync.Mutex

	// channel
	ChanSim        chan interface{}
	ChanBlockchain chan interface{}

	// trust value storage and misbehaving flag results for epochs
	// each experiment instance has its own trust value storage
	TrustValueStorageHead *map[string]*dtmtype.TrustValueStorageHead

	// a random source
	R *randutil.RandUtil
}

func PrepareDTMLogicModuleSession(
	simCfg *config.SimConfig, expCfg *map[string]*config.ExperimentConfig,
	c chan interface{}) *DTMLogicSession {

	dtmSession := &DTMLogicSession{}

	// init the simulator config and experiment config
	dtmSession.SimConfig = simCfg
	dtmSession.Config = expCfg

	// inter module communication
	dtmSession.ChanSim = c

	// random source
	dtmSession.R = randutil.InitRand(123)

	// start to parse the experiment configuration and init the storage area
	tmp := make(map[string]*dtmtype.TrustValueStorageHead)
	dtmSession.TrustValueStorageHead = &tmp
	for expName := range *expCfg {
		(*dtmSession.TrustValueStorageHead)[expName] = dtmtype.InitTrustValueStorage()
	}

	return dtmSession
}
