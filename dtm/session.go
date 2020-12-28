package dtm

import (
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/config"
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
	CompromisedRSUBitMap *sync.Map // only valid for specific epoch

	// pointer to the vehicles and RSU
	// TODO: I don't know if it is a good idea to use mutex via pointer
	Vehicles *[]*vehicle.Vehicle
	RSUs     *[][]*RSU
	vmu      *sync.Mutex
	rmu      *sync.Mutex

	// channel
	ChanSim        chan interface{}
	ChanBlockchain chan interface{}

	// trust value storage for each setups
	// each experiment instance has its own trust value storage
	TrustValueStorageHead *map[string]*dtmtype.TrustValueStorageHead

	// a random source
	R *randutil.RandUtil
}
