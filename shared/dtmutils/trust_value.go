package dtmutils

type TrustValueOffset struct {
	VehicleId uint64
	Slot uint64
	TrustValueOffset float32
}

type TrustValue struct {
	VehicleId uint64
	Epoch uint64
	TrustValue float32
}

type TrustValueRecord struct {
	Epoch uint64
	// [vehicleId<uint64>]TrustValue<float32>
	TrustValueList *map[uint64]float32
	pNext *TrustValueRecord
	pPrevious *TrustValueRecord
}