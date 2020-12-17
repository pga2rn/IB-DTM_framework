package dtmutil

import "sync"

// trust value starts range from -1 ~ 1
// starts from 0

type TrustValueOffset struct {
	VehicleId uint64
	Slot uint64
	TrustValueOffset float32
	Weight float32
}

type TrustValue struct {
	VehicleId uint64
	Epoch uint64
	TrustValue float32
}

const (
	Rountine = 0.5
	Crital = 0.7
	Fatal = 0.9
)

type TrustValueStorageHead struct {
	p* TrustValueStorage
	mu sync.Mutex
}

// data structure to hold every vehicle's trust value of specific epoch
type TrustValueStorage struct {
	Epoch uint64
	// [vehicleId<uint64>]TrustValue<float32>
	TrustValueList *map[uint64]float32
	pNext *TrustValueStorage
	pPrevious *TrustValueStorage
}