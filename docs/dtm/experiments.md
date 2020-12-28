# Experiments

## Experiment variables

### Misbehaving vehicles

1. portion of misbehaving vehicles
2. weight assignment for each trust value offsets

### Compromised RSU

1. portion of compromised RSU
2. types of evil that RSU can do

### Time factor

1. time factor function choice

### trace back epochs

1. how many previous epochs' trust value will also be taken account.

## Experiment scenario

### Baseline 0(no compromised RSU)

Simulated the pure distributed trust management, RSU collects trust value offsets among the network, and calculate trust value from the trust value offsets.

Assuming that all RSU in the network can get all the trust value offsets on time, and the storage & broadcasting of trust value offsets will not be dropped or alter. Of course if we take the alteration of trust value offsets when broadcasting, the results will only be worse, the baseline is the very ideal scenario of infrastructure-assisted distributed trust management/

1. RSU :No compromised RSU

2. Vehicles: With misbehaving vehicles

3. Trust value generation
   1. per epoch calculation, without time factor
   2. without bias(no compromised RSU)

### Baseline 1(with compromised RSU)

1. RSU: with compromised RSU
2. Vehicles: with misbehaving vehicles
3. Trust value  generation: 
   1. per epoch calculation, without time factor
   2. with bias(compromised RSU will do evils)

### Baseline 2 (trust value calculate with previous epochs data)

1. RSU: with compromised RSU
2. Vehicles: with misbehaving vehicles
3. Trust value generation:
   1. from previous N epochs, without time factor
   2. with bias(compromised RSU will do evils)

### Proposal 0(no compromised RSU)

1. RSU: no compromised RSU
2. Vehicles: with misbehaving vehicles
3. Trust value offsets generation & broadcast: via blockchain, with proposed consensus logic
4. Trust value generation:
   1. from previous N epochs, with time factor
   2. without bias

### Proposal 1(with compromised RSU)

The same as proposal 0, but with bias

## Experiment result

The simulator will generate misbehaving vehicles assignment at the beginning of simulation, that is the correct answer.

Each experiment scenario will also flag out misbehaving vehicles, the misbehaving vehicles that the solution points out is the solutions' answer.

## Experiment metrics

When we collect the raw experiment results, we can use the following metrics to evaluate them:

1. plain bias between the trust values
2. results of identifying misbehaving vehicles
   1. basic metrics: TP, FP, FN, TN
   2. advanced metrics: Recall, Precision, F1score, Accuracy

