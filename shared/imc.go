package shared

import (
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/rsu"
	"github.com/pga2rn/ib-dtm_framework/vehicle"
	"sync"
)

// struct that used for communication between simulator module and DTM module
type SimDTMCommunication struct {
	Slot                 uint64
	CompromisedRSUBitMap *bitmap.Threadsafe // only pass the pointer
}

// struct for initializing the dtm
type SimInitDTMCommunication struct {
	MisbehavingVehicleBitMap *bitmap.Threadsafe
	Vehicles                 *[]*vehicle.Vehicle
	RSUs                     *[][]*rsu.RSU
	Vmu                      *sync.Mutex
	Rmu                      *sync.Mutex
}

// struct for dtm module and blockchain module
// TODO: definition for dtm and blockchain module communication
type DTMBlockchainCommunication struct {
}
