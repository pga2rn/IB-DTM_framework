package dtmtype

import "sync"

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

// link list head of trustvaluestorage list
type TrustValueStorageHead struct {
	HeadEpoch  uint64 `the epoch of current head`
	EpochCount int    `total epoch being recorded`
	PtrNext    *TrustValueStorage
}

// data structure to hold every vehicle's trust value of specific epoch
type TrustValueStorage struct {
	Epoch          uint64
	TrustValueList *sync.Map `map[uint64]float32`
	PtrNext        *TrustValueStorage
	PtrPrevious    *TrustValueStorage
}

func InitTrustValueStorage() *TrustValueStorageHead {
	return &TrustValueStorageHead{
		HeadEpoch:  0,
		EpochCount: -1,
	}
}

// TODO: implement the link list for trustvaluestorage
func InitTrustValueStorageObject(epoch uint64) *TrustValueStorage {
	storage := TrustValueStorage{}
	storage.Epoch = epoch
	storage.TrustValueList = &sync.Map{}
	return &storage
}
