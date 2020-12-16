package statistics

// VehicleRecords records all data of vehicles
type VehicleRecords struct {
	epoch          uint32
	vehicleList    []*Vehicle
	trustValueList []*TrustValue
}

// MapRecords records all data of a map
type MapRecords struct {
	epoch uint32
}

// maintain a linked list of records per epoch