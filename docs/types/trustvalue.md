# Trust value

The data structures to store trust value and trust value offsets are as follow:

### trust value & trust value offset

```go
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
```

### trust value storage

```go
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
```

