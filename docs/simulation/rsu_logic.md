# RSU logic

This document describes how RSU behaves in the simulation.

## Management Zone & sensing zone

Each RSU has it **management zone** and **sensing zone**.

**Management zone** is the area that RSU can control and manage the vehicles, it is one-to-one mapping to the **cross** in the sim-map. Trust value offsets of vehicles entered the specific cross will be gather by the **RSU** at the cross.

**Sensing zone** is the surrounding area that RSU aware of, typically the closest 4 RSUs near the specific RSU. When other RSU validating the trust value offsets reported by the specific RSU, validator RSU can query from the RSU within the sensing zone to gather information.

## Compromised RSU

The RSUs will be assigned to be **normal** or **compromised** at the start of each epoch, the compromised RSU will do the following 3 types of evil:

```go
const (
	FlipTrustValueOffset = iota
	DropPositiveTrustValueOffset
	ForgeTrustValueOffset
)
```

1. FlipTrustValueOffset: the compromised RSU will report the reversed trust value offsets,
2. DropTrustValueOffset: the compromised RSU will intendedly ignore some trust value offsets,
3. ForgeTrustValueOffset: the compromised RSU will report forged trust value offsets for vehicles that not in the managed zone.

## Distributed trust management functionality

In the simulation, the DTM logic that performed by RSU is done by the simulator, including:

1. trust ratings gathering: the simulator will generate trust value offsets directly, by pass the gathering of trust ratings;
2. trust value offsets generation: the same as 1, trust value offsets for specific vehicles at specific slot are generated directly and dispatch to each RSU instance by the simulator;
3. trust value generation: 