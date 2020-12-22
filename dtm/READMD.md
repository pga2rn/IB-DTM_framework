# Distributed trust management

The dtm subfolder contains the data structure used for the simulation.

## RSU

rsu.go defines the data structure of RSU.

### Properties

the following properties describe a rsu instance:

1. Id: unique Id to identify the RSU, it is one-to-one mapping to the crosses in the map
2. Position: indicate the position of the RSU in the map, one-to-one mapping to the cross
3. Epoch/Slot: syncing with the time stream
4. TrustValueOffsetPerSlot:  stores the vehicles' trust value offsets within an epoch

### Compromised RSU

Some of the RSU will be assigned to be the so-called **compromised RSU** at the beginning of an epoch, they will do evil relating to the trust value offsets, detailed can be founded at core/RSU_LOGIC.md.