package dtmtype

import (
	"errors"
	"sync"
)

type TrustValue struct {
	VehicleId  uint64
	Epoch      uint64
	TrustValue float32
}

// thread safe map
type TrustValuesPerEpoch = sync.Map // map[<vehicleId>uint64]float32

// link list head of trust value storage list
type TrustValueStorageHead struct {
	mu         sync.Mutex
	headEpoch  uint64 `the epoch of current head`
	headPtr    *TrustValueStorage
	epochCount int `total epoch being recorded`
	ptrNext    *TrustValueStorage
}

// data structure to hold every vehicle's trust value of specific epoch
type TrustValueStorage struct {
	epoch          uint64
	trustValueList *TrustValuesPerEpoch
	ptrNext        *TrustValueStorage
	ptrPrevious    *TrustValueStorage
}

// constructor of trust value storage
func InitTrustValueStorage() *TrustValueStorageHead {
	return &TrustValueStorageHead{
		mu:         sync.Mutex{},
		headEpoch:  -1,
		epochCount: 0,
	}
}

// init a storage for specific epoch
func (head *TrustValueStorageHead) InitTrustValueStorageObject(epoch uint64) (*TrustValueStorage, error) {
	if epoch != (head.headEpoch + 1){
		return nil, errors.New("storage is out of sync with the simulation")
	}

	// init the new storage object
	storage := &TrustValueStorage{
		epoch:       epoch,
		trustValueList: nil,
		ptrNext:     nil,
		ptrPrevious: head.headPtr,
	}

	head.mu.Lock()
	// update the head block
	head.headPtr.ptrNext = storage

	// update the head information in the head
	head.headPtr, head.headEpoch, head.epochCount = storage, epoch, head.epochCount+1
	head.mu.Unlock()

	return storage, nil
}

func (head *TrustValueStorageHead) GetEpochInformation() (uint64, int){
	return head.headEpoch, head.epochCount
}

func (head *TrustValueStorageHead) GetTrustValueStorageForEpoch (epoch uint64) *TrustValueStorage{
	if epoch > head.headEpoch {
		return nil
	}

	ptr := head.ptrNext
	for i := uint64(0); i < epoch; i++{
		ptr = ptr.ptrNext
	}
	return ptr
}

// assign trust value list to a storage object
func (storage *TrustValueStorage) SetTrustValueList(epoch uint64, list *TrustValuesPerEpoch) error{
	if storage.epoch != epoch {
		return errors.New("mismatch input epoch and storage epoch")
	}
	storage.trustValueList = list
	return nil
}

func (storage *TrustValueStorage) GetTrustValueList() (uint64, *TrustValuesPerEpoch){
	return storage.epoch, storage.trustValueList
}