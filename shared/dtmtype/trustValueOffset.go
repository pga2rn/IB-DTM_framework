package dtmtype

import "sync"

// trust value starts range from -1 ~ 1
// starts from 0

type TrustValueOffset struct {
	VehicleId        uint64
	Slot             uint64
	TrustValueOffset float32
	Weight           float32
}

// we use sync.map for thread safe
type TrustValueOffsetsPerSlot = sync.Map // map[<vehicleId>uint64]*TrustValueOffset
//type TrustValueOffsetsPerEpoch = sync.Map // map[<slot>uint64]*TrustValueOffsetsPerSlot

// trust value offset weight
const (
	Routine  = 0.5
	Critical = 0.7
	Fatal    = 0.9
)

// sync.map can be used directly without extra initializing