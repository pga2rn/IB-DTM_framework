package shared

import (
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/rsu"
	"github.com/pga2rn/ib-dtm_framework/vehicle"
	"sync"
)

// struct that used for communication between simulator module and DTM module
type SimDTMEpochCommunication struct {
	Slot                 uint32
	CompromisedRSUBitMap *bitmap.Threadsafe // only pass the pointer
}

// for proposal
type SimDTMSlotCommunication struct {
	Slot uint32
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

// structs for rpc server
type DTMRPCCommunication struct {
	Epoch  uint32                     `protobuf:"varint,1,opt,name=epoch,proto3" json:"epoch,omitempty"`
	Bundle []*StatisticsPerExperiment `protobuf:"bytes,2,rep,name=bundle,proto3" json:"bundle,omitempty"`
}

type StatisticsPerExperiment struct {
	// experiment name & type
	Name string         `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Type ExperimentType `protobuf:"varint,2,opt,name=type,proto3,enum=rpc.pb.ExperimentType" json:"type,omitempty"` // 0: baseline, 1: proposal
	// concrete experiment results
	Epoch     uint32  `protobuf:"varint,3,opt,name=epoch,proto3" json:"epoch,omitempty"`
	Tp        float32 `protobuf:"fixed32,10,opt,name=tp,proto3" json:"tp,omitempty"`
	Fp        float32 `protobuf:"fixed32,11,opt,name=fp,proto3" json:"fp,omitempty"`
	Tn        float32 `protobuf:"fixed32,12,opt,name=tn,proto3" json:"tn,omitempty"`
	Fn        float32 `protobuf:"fixed32,13,opt,name=fn,proto3" json:"fn,omitempty"`
	Recall    float32 `protobuf:"fixed32,14,opt,name=recall,proto3" json:"recall,omitempty"`
	Precision float32 `protobuf:"fixed32,15,opt,name=precision,proto3" json:"precision,omitempty"`
	F1Score   float32 `protobuf:"fixed32,16,opt,name=f1score,proto3" json:"f1score,omitempty"`
	Acc       float32 `protobuf:"fixed32,17,opt,name=acc,proto3" json:"acc,omitempty"`
}

type ExperimentType = int32

const (
	BASELINE = iota
	PROPOSAL
)
