package shared

import (
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/rpc/pb"
	"github.com/pga2rn/ib-dtm_framework/rsu"
	"github.com/pga2rn/ib-dtm_framework/shared/fwtype"
	"sync"
)

// struct that used for communication between simulator module and DTM module
type SimDTMEpochCommunication struct {
	Slot                 uint32
	ActiveVehiclesNum    int32
	CompromisedRSUBitMap *bitmap.Threadsafe // only pass the pointer
}

// directly use protobuff definition
//type StatisticsPerExperiment
//type StatisticsBundle

// dtm ~ ibdtm communication
type IBDTM2DTMCommunication struct {
	Epoch          uint32
	ExpName        string
	TrustValueList *fwtype.TrustValuesPerEpoch
}

// struct for initializing the dtm & ib-dtm module
type SimInitDTMCommunication struct {
	MisbehavingVehicleBitMap *bitmap.Threadsafe
	RSUs                     *[][]*rsu.RSU
	Rmu                      *sync.Mutex
}
type SimInitIBDTMCommunication = SimInitDTMCommunication

// structs for rpc server
type DTMRPCCommunication = pb.StatisticsBundle
