package fwtype

import (
	"errors"
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/rpc/pb"
	"sync"
)

// thread safe map
type TrustValuesPerEpoch = sync.Map // map[<vehicleId>uint32]float32

// link list head of trust value storage list
type TrustValueStorageHead struct {
	mu         sync.Mutex
	headEpoch  uint32 `the epoch of current head`
	headPtr    *TrustValueStorage
	epochCount int `total epoch being recorded`
	ptrNext    *TrustValueStorage
}

// data structure to hold every vehicle's trust value of specific epoch
type TrustValueStorage struct {
	epoch                    uint32
	trustValueList           *TrustValuesPerEpoch
	misbehavingVehicleBitMap *bitmap.Threadsafe
	statisticsPack           *pb.StatisticsPerExperiment
	ptrNext                  *TrustValueStorage
	ptrPrevious              *TrustValueStorage
}

// constructor of trust value storage
func InitTrustValueStorage() *TrustValueStorageHead {
	return &TrustValueStorageHead{
		mu:         sync.Mutex{},
		headEpoch:  0,
		epochCount: 0,
	}
}

// init a storage for specific epoch, the way to add a new block into the linked list
// and then we can attach the trust value list to the returned new block via SetTrustValueList
func (head *TrustValueStorageHead) InitTrustValueStorageObject(epoch uint32, cfg *config.SimConfig) (*TrustValueStorage, error) {
	if epoch != (head.headEpoch+1) && epoch != 0 {
		return nil, errors.New("storage is out of sync with the simulation")
	}

	// init the new storage object
	storage := &TrustValueStorage{
		epoch:                    epoch,
		trustValueList:           &TrustValuesPerEpoch{},
		misbehavingVehicleBitMap: bitmap.NewTS(cfg.VehicleNumMax),
		ptrNext:                  nil,
		ptrPrevious:              head.headPtr,
	}

	head.mu.Lock()
	// update the head block
	if head.headPtr != nil {
		head.headPtr.ptrNext = storage
		head.headPtr = storage
	} else {
		// for slot 0
		head.ptrNext = storage
		head.headPtr = storage
	}

	// update the head information in the head
	head.headEpoch, head.epochCount = epoch, head.epochCount+1
	head.mu.Unlock()

	return storage, nil
}

func (head *TrustValueStorageHead) GetEpochInformation() (uint32, int) {
	return head.headEpoch, head.epochCount
}

func (head *TrustValueStorageHead) GetHeadBlock() *TrustValueStorage {
	return head.headPtr
}

func (head *TrustValueStorageHead) GetTrustValueStorageForEpoch(epoch uint32) *TrustValueStorage {
	if epoch > head.headEpoch {
		return nil
	}

	ptr := head.ptrNext
	for i := uint32(0); i < epoch; i++ {
		ptr = ptr.ptrNext
	}
	return ptr
}

func (storage *TrustValueStorage) AddTrustRatingForVehicle(vid uint32, v float32) {
	list := storage.trustValueList
	if op, ok := list.LoadOrStore(vid, v); ok {
		list.Store(vid, v+op.(float32))
	}
}
func (storage *TrustValueStorage) GetTrustRatingForVehicle(vid uint32) (float32, bool) {
	list := storage.trustValueList
	if res, ok := list.Load(vid); ok {
		return res.(float32), true
	}
	return 0, false
}

// assign trust value list to a storage object
func (storage *TrustValueStorage) SetTrustValueList(epoch uint32, list *TrustValuesPerEpoch) error {
	if storage.epoch != epoch {
		return errors.New("mismatch input epoch and storage epoch")
	}
	storage.trustValueList = list
	return nil
}

func (storage *TrustValueStorage) GetTrustValueList() (uint32, *TrustValuesPerEpoch, *bitmap.Threadsafe) {
	return storage.epoch, storage.trustValueList, storage.misbehavingVehicleBitMap
}

func (storage *TrustValueStorage) SetStatistics(bundle *pb.StatisticsPerExperiment) {
	storage.statisticsPack = bundle
}

func (storage *TrustValueStorage) GetStatistics() *pb.StatisticsPerExperiment {
	return storage.statisticsPack
}
