# Distributed trust management simulation logic

core/dtm_logic.go implements the distributed strust management logic for the simulation.

## Routine

The distributed trust management is being simulated as follow:

1. at every slot, the simulator will randomly assign each vehicle with **trust value offsets** with **weight**, and then attach these offsets to each RSU according to which cross is the vehicle at;
2. at the end of epoch,  the simulator will gather all RSUs' trust value offsets storage, and then calculate trust value with ***weighted* trust value offsets **and **time factor**.

## Trust measurement over vehicle

The trust management measure the trustworthiness of vehicle by trust value. Trust value is generated from trust value offsets at the end of each epoch, and trust value offsets are generated and assigned to each vehicle at every slot.

### Trust value offset

Trust value offset represents the vehicle's trustworthiness at a specific time slot, it also attached with a weight which related to the type of message as follow:

```go
// defined in shared/dtmtype/trustvalue.go
const (
	Rountine = 0.5 // routinue message, like position and status broadcasting
	Crital   = 0.7 // critial on-road message, like traffic volume or normal event
	Fatal    = 0.9 // fatal on-road message, like traffic accident
)
```

### Trust value

Trust value represents the on-going vehicle's trustworthiness for a specific epoch. It is the sum of tuned trust value offsets from previous $N$ epochs.

RSUs query all trust value offsets of specific vehicles within an epoch. Trust value $V_i$ of vehicle $i$ in epoch $E$  can be calculated as follow, $S$ is the number of slots in $E$, $v_s$ is the sum of  trust value offsets for slot $s$, function $f$ is the time factor function, taking the start time $t_s$ of slot $s$ as parameter. 

$$
V_i=\sum^{S}_{s=0}{v_s \times f(t_s)},\quad V_i \in [-1, 1]
$$

## RSU behavior

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

## Generation logic

### Trust value offsets generation

At every slot, the simulator will iterate through all the vehicles and:

1. check the ```MisbehavingVehicleBitMap``` in **simulation session** to see the vehicle is bad or not,
   1. if it is good vehicle, the simulator will rate the vehicle with +1 raw trust value offset
   2. if it is bad vehicle, there is still $1\%$ chance that the vehicle will not do evil, otherwise assign -1 raw trust value offset
2. randomly assign the rating with a message type (weight), the possibility of choices is the same as the weight for convenience,
3. save the trust value offsets to the RSU that vehicle located at. 

### Trust value generation

At the end of each epoch, the simulator will iterate through all the RSUs, collecting trust value offsets, and then calculate the trust value from the sum of offsets,

1. check the ```CompromisedRSUBitMap``` to see whether the RSU is compromised or not,
   1. if not compromised, the RSU will report true trust value offsets,
   2. if compromised, it will randomly choose from ```FlipTrustValueOffset``` and ```DropPositiveTrustValueOffset``` and report incorrect trust value offsets
   3. if compromised, it will also ```ForgeTrustValueOffest```, report trust value offsets that not stored in the RSU data field,
2. gather all the trust value offsets from RSUs, and then calculate the sum for each vehicle and generate trust value for it.