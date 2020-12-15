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