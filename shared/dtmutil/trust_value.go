package dtmutil

import (
	"sync"
)

// TODO: implement mutex for trustvaluestorage

// trust value starts range from -1 ~ 1
// starts from 0

type TrustValueOffset struct {
	VehicleId        uint64
	Slot             uint64
	TrustValueOffset float32
	Weight           float32
}

type TrustValue struct {
	VehicleId  uint64
	Epoch      uint64
	TrustValue float32
}

const (
	Rountine = 0.5
	Crital   = 0.7
	Fatal    = 0.9
)

// data structure to hold every vehicle's trust value of specific epoch
type TrustValueStorage struct {
	Epoch uint64
	Mu *sync.Mutex // thread safe
	// [vehicleId<uint64>]TrustValue<float32>
	TrustValueList *map[uint64]float32
	pNext          *TrustValueStorage
	pPrevious      *TrustValueStorage
}

// TODO: implement the link list for trustvaluestorage
func InitTrustValueStorageObject(epoch uint64) *TrustValueStorage{
	storage := TrustValueStorage{}
	storage.Epoch = epoch
	storage.mu = &sync.Mutex{}
	storage.TrustValueList = func() *map[uint64]float32{ s := make(map[uint64]float32); return &s}()

	return &storage
}