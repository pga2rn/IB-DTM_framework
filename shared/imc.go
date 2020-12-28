package shared

import (
	"github.com/pga2rn/ib-dtm_framework/dtm"
	"github.com/pga2rn/ib-dtm_framework/vehicle"
	"sync"
)

// struct that used for communication between simulator module and DTM module
type SimDTMCommunication struct {
	Slot                 uint64
	CompromisedRSUBitMap *sync.Map // only pass the pointer
}

// struct for initializing the dtm
type SimInitDTMCommunication struct {
	Vehicles *[]*vehicle.Vehicle
	RSUs     *[][]*dtm.RSU
	vmu      *sync.Mutex
	rmu      *sync.Mutex
}

// struct for dtm module and blockchain module
// TODO: definition for dtm and blockchain module communication
type DTMBlockchainCommunication struct {
}