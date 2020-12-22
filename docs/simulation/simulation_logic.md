# Simulation logic

The simulation framework implements a N * M grid, each cross in the map represents a **RSU**. The simulated **vehicles** will move from one cross to another cross, enter or leave the map, from time to time. The RSU collects the **trust value offset** of each vehicle, and further calculate the **trust value** of each vehicle.

The time stream of simulation is divided into many **epochs**, each epoch contains several **slots** which lasts for many seconds. The whole map will update at every slot, move vehicles and generate trust value offsets for each vehicle. At the end of each epoch, the trust value of each vehicles will be generated.

## Simulation variables

1. Misbehaving vehicles

   The assigned misbehaving vehicles' **trust value offsets *Raw Value*** will always be -1. The assignment will only be done every $N$ epochs.

2. Compromised RSU

   The assigned compromised RSU will alter the trust value offsets when generating trust value at the end of an epoch. The assignment of compromised RSU will be regenerated very $N$ epochs.

## Simulation set up

Before the simulation can be fired up, the following data structure will be initialized:

1. Simulation session
2. Simulation map
3. RSU
4. Vehicles

And at the beginning of **genesis**, the following simulation variables will be initialized:

1. Compromised RSU assignment
2. Misbehaving vehicles assignment

## Simulation routine

### Ticker

The simulation is fully sync, there is a time ticker in the background counting the beat. The ticker will ticket at every slot, with the interval of ```SecondsPerSlot``` seconds. 

### Process Epoch

At the end of an epoch,  and also the start of the next epoch, the following routines will be executed:

1. calculate the trust value of vehicles for the epoch
2. reassign the compromised RSU
3. clean up RSUs storage fields for storing trust value offsets for new epoch

Accordingly, the following functions will be executed:

```go
genTrustValue()
initAssignCompromisedRSU()
```

### Process Slot

During every slot, the map will be updated in the following routine:

1. move the vehicles within the map
2. generate trust value offsets for every active vehicles within the map

Accordingly, the following functions will be executed:

```go
// move the vehicles within the map
moveVehiclesPerSlot()
// execute distributed trust management logic
executeDTMLogicPerSlot()
```

