# Session

The ```session``` type defines a simulation session, it roughly contains the following parts:

```go
type SimulationSession struct {
	// simulation session config
    // simulation map
    // time related
    // simulation status
    // simulation objects storage
    // Random source
}
```

## Session config

Please refer to docs/type/config.md for detail.

```go
	// config of the current simulation session
	Config *config.Config
```

## Simulation Map

The ```Map``` pointer points to the map instance that the simulation runs on, please refer to docs/type/sim-map.md for detail.

```go
	// pointer to the map
	Map *simmap.Map
```

## Time related

```go
	// time
	Ticker timeutil.Ticker
	// epoch and slot stored in session should only be used when gathering reports
	Epoch uint64
	Slot  uint64
```

The ```Ticker``` is a ```timeutil.Ticker``` that emit a pulse at the start of every slot, to 'push' the process of simulation. The ```Epoch``` and ```Slot``` indicates where we are right now in the time stream at the simulation.

## Status tracking

Status section keep the record of current status of the simulation.

```go
	// current status
	// vehicle
	ActiveVehiclesNum         int
	ActiveVehiclesBitMap      *bitmap.Threadsafe
	MisbehaviorVehicleBitMap  *bitmap.Threadsafe
	MisbehaviorVehiclePortion float32
	// RSU
	CompromisedRSUPortion float32
	// store the ID(index) of compromised RSU of this slot
	CompromisedRSUBitMap *bitmap.Threadsafe
	// a complete list that stores every vehicle's trust value
	TrustValueList *sync.Map // without bias
	BiasedTrustValueList   *sync.Map // with bias
```

## Simulated objects

This section stores the data of instances of vehicles and RSUs in the simulation.

```go
	// a list of all vehicles in the map
	Vehicles []*vehicle.Vehicle
	// a 2d array store the RSU data structure
	// aligned with the map structure
	RSUs [][]*dtm.RSU
```

## Random source

All the ```Random``` operations in the simulation session use ```R``` as random source.

```go
	// a random generator, for determined random
	R *rand.Rand
```