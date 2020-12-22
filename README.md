# IB_DTM-framework

IB_DTM-framework implements a simulation framework for testing distributed trust management application.

## Overview

The simulation framework implements a N * M map, each cross in the map represents a RSU. The simulated vehicles will be moved from one cross to another cross, enter or leave the map, from time to time. The RSU collects the trust value offset of each vehicle, and further calculate the trust value of each vehicle.

The time stream of simulation is divided into many epochs, each epoch contains several slots which lasts for many seconds. The whole map will update at every slot, move vehicles and generate trust value offsets for each vehicle. At the end of each epoch, the trust value of each vehicles will be generated.

## Documents

There are also documents in the subfolder, including:

1. core/ : logics of simulation
2. shared/timefactor: definition of timefactor
3. shared/dtmtype: definition of data structure for holding trust value, trust value offsets and trust value storage
4. sim-map/: definition and construction of the simulation map
5. services: components of the simulator
6. statistics:

## project Architecture:

1. config

   config defines the available configurations for the simulation

2. core

   core implements the core logics of simulation, including distributed trust management logic, rsu logics, vehicle movement logic, etc.

3. dtm

   dtm defines the data structure needed for distributed trust managment, mainly the RSU.

4. rpc

   rpc implements the RPC interface of the simulator, mainly for data query and debug.

5. service

   service inits and fires up each components of the simulator

6. shared

   shared implements some shared utils and data types for the simulator

7. sim-map

   sim-map defines the map that simulation runs on.

8. vehicle

   vehicle defines the data structure for individual vehicles

## Usage

1. install and config the go environment, version 1.15.* is recommended.

2. check out the config/ folder for configuration and apply changes to it as needed, available configurations are as follow:

   ```go
   type Config struct {
       // map size
   	XLen int
   	YLen int
   
   	// simulation config
       // range of number of simulated vehicles allowd in the map, 
   	VehicleNumMax              int
   	VehicleNumMin              int
       // range of portion of compromised vehicle allowd in the map,
       // applied once for a simulation session
   	MisbehaveVehiclePortionMax float32
   	MisbehaveVehiclePortionMin float32
       // number of RSU in a simulation, should be the same as totall crosses of the map 
   	RSUNum                   int // XLen * YLen
       // portion of the compromised RSU, applied at every epoch
   	CompromisedRSUPortionMax float32 // from 0 ~ 1
   	CompromisedRSUPortionMin float32 // from 0 ~ 1
       // the timefactor type for tunning the trust value offset raw value
       // available type: Exp, Linear, Power, Sin, Log
       // see shared/timefactor/README.md for detailed
   	TimeFactorType int
   
   	// time config
       // the kick start time of the simulation
   	Genesis           time.Time
   	SlotsPerEpoch     uint64
   	SecondsPerSlot    uint64 // in seconds
   }
   ```

3. cd into the IB_DTM-framework folder, execute the following commands:

   ```shell
   $ go run ./main.go
   ```

# Experiment

## Variables

The variables that the simulations will use are as follow:

1. misbehaving vehicles

   portion of the vehicles will be assigned as misbehaving vehicles for a simulation session, their trust value offsets will always be -1 during the whole simulation

2. compromised RSU

   portion of RSU will be assigned as compromised RSU, they will alter the trust value offsets when generating the trust value

3. weight of trust value offsets

   each rating (trust value offsets) will be randomly assigned a weight,  including

   ```go
   const (
   	Rountine = 0.5 // routinue message, like position and status broadcasting
   	Crital   = 0.7 // critial on-road message, like traffic volume or normal event
   	Fatal    = 0.9 // fatal on-road message, like traffic accident
   )
   ```

4. timefactor

   the tuning factor applied to the raw trust value offsets when calculating trust value, see shared/timefactor/README.md for detailed. 

## Metrics

Metrics we use to evaluate are as follow:

1. bias: plain trust value bias caused by compromised RSU
2. TP, FP, FN, TN
3. Recall, Precision, Accuracy