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
	mu        sync.Mutex
	headEpoch uint32
	headPtr   *TrustValueStorage
}

// data structure to hold every vehicle's trust value of specific epoch
type TrustValueStorage struct {
	epoch                    uint32
	trustValueList           *TrustValuesPerEpoch
	misbehavingVehicleBitMap *bitmap.Threadsafe
	statisticsPack           *pb.StatisticsPerExperiment
}

// constructor of trust value storage
func InitTrustValueStorage() *TrustValueStorageHead {
	return &TrustValueStorageHead{
		mu:        sync.Mutex{},
		headEpoch: 0,
	}
}

// init a new storage object
func (head *TrustValueStorageHead) InitTrustValueStorageObject(epoch uint32, cfg *config.SimConfig) (*TrustValueStorage, error) {
	if epoch != (head.headEpoch+1) && epoch != 0 {
		return nil, errors.New("storage is out of sync with the simulation")
	}

	// init the new storage object
	storage := &TrustValueStorage{
		epoch:                    epoch,
		trustValueList:           &TrustValuesPerEpoch{},
		misbehavingVehicleBitMap: bitmap.NewTS(cfg.VehicleNumMax),
	}

	head.mu.Lock()
	// record the results
	head.headPtr, head.headEpoch = storage, epoch
	head.mu.Unlock()

	return storage, nil
}

func (head *TrustValueStorageHead) GetEpochInformation() uint32 {
	return head.headEpoch
}

func (head *TrustValueStorageHead) GetHeadBlock() *TrustValueStorage {
	return head.headPtr
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
