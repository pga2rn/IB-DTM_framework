package dtm

import "github.com/pga2rn/ib-dtm_framework/simulator/core"

type TrustValueOffset struct {
	VehicleId uint64

	TimeStamp core.Beacon

	TrustValueOffset float32
}

type TrustValue struct {
	VehicleId uint64

	Epoch uint64

	TrustValue float32
}