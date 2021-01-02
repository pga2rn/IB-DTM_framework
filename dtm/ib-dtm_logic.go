package dtm

import (
	"github.com/boljen/go-bitmap"
	"sync"
)

// struct for proposal experiment setups
type IBDTMStorage struct {
	// flag out compromised RSU
	flaggedOutCompromisedRSUBitmapPerEpoch *sync.Map // map[<epoch>uint32]*bitmap.threadsafe
}

func (storage *IBDTMStorage) FlagRSUForEpoch(epoch, rid uint32) bool {
	if bmap, ok := storage.flaggedOutCompromisedRSUBitmapPerEpoch.Load(epoch); ok {
		bmap.(*bitmap.Threadsafe).Set(int(rid), true)
		return ok
	} else {
		return ok
	}
}

func (storage *IBDTMStorage) GetFlaggedBitmapForEpoch(epoch uint32) (*bitmap.Threadsafe, bool) {
	if bmap, ok := storage.flaggedOutCompromisedRSUBitmapPerEpoch.Load(epoch); ok {
		return bmap.(*bitmap.Threadsafe), ok
	} else {
		return nil, false
	}
}

func InitIBDTMStorageMap() *map[string]*IBDTMStorage {
	tmp := make(map[string]*IBDTMStorage)
	return &tmp
}

func InitIBDTMStorage() *IBDTMStorage {
	return &IBDTMStorage{
		flaggedOutCompromisedRSUBitmapPerEpoch: &sync.Map{},
	}
}
