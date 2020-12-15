package dtmutils

// definition of a single trust value offset record
type TrustValueOffset struct {
	VehicleId uint64

	Epoch uint64
	Slot uint64

	TrustValueOffset float32
}