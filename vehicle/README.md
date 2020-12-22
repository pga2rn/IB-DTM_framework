# Vehicles

Vehicle, together with RSU,  is the core elements in the simulation. During the simulation, at every time slot, vehicle will be move from one cross to another cross within the map, or leave the map.

## Properties

Each vehicle is being described as follow:

1. Id: unique id to identify the each vehicle
2. Position: indicate the location of the vehicle right now
3. VehicleStatus: active, if the vehicle is on the map, or otherwise, inactive
4. LastMovementDirection: record the direction of the vehicle's last maneuvering

## Moving Logic

In order to make the movement of the vehicle more natural, and let the vehicle stays in the map as long as possible, the following rules are applied when moving the vehicle:

1. the vehicle will try to moving toward the center of the map
2. the vehicle will try to not change the direction too much

Detail of the algorithm can be founded at core/VEHICLE.md