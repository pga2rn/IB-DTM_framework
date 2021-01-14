# IB-DTM logic

## experiment settings

vehicles: 4096

epoch size: 16 slots

validators number: 16*16 = 256 RSUs

committee size:   validators number / epoch size = 16

uploader cover ratio: 0.5

shard number: 16 * uploader cover ratio = 8

## proof of stake

2 types of stake: effective stake & its stake

effective stake: 

​	initial stake = vehicles * 1.5 / RSU = 24, assigned to each RSU at the genesis

​	effective stake upper bound: 32

​	exit threshold: 0.5 * effective stake = 16

its stake:

​	unique vehicles that being witnessed within trackback epochs



factor: its stake / vehicles



vote weight is decided by the effective stake:

over 2/3 of total stake must be gained to approve a block

## reward

base reward = effective * / sqrt(vehicles) * uploader cover ratio = 1



at every end of epoch:

reward = base reward * sqrt(factor)



## penalty

if slashed: directly kick from the network

if the upload block has been rejected due to unable to gain enough votes:

penalty_factor = 3

penalty = base reward * (sqrt(factor) or 1)* penalty_factor



## life cycle

if the effective stake lower than 0.5 * effective stake upper bound, the RSU will be kick out of the consensus



## consensus

at the beginning of the epoch, checkpoint

the assignment of proposer & committee will be decided with consensus randomly



# Implementation detail

## blockchain definition

1. blockchain status

    ```go
type BlockchainStatus struct {
        cofig simconfig
        slot uint32
    
        validatorbitmap *bitmap
        validatorlist []*validator
        balance []*balance
    
        // per epoch slashings, will be counted at the end of epoch
        // collected from the uploaded blocks
        slashings []*slashing
}
    ```

2. blockchain block

   ```go
   type BlockchainBlock struct {
       slot uint32
    validatorId uint32
   	slashings []*slashing
       
       votes []*bitmap
   }
   ```
   
   

## life cycle logic

1. at the beginning of genesis, all validators will be assigned a initial stake
2. each epoch, validator will receive penalties or rewards accordingly
3. if effective stake lower than 0.5 * upper bound, the validator will be kicked
4. if the validator has been slashed, it will be kicked



## committee generation

```python
def compute_proposer_index(state: BeaconState, indices: Sequence[ValidatorIndex], seed: Bytes32) -> ValidatorIndex:
    """
    Return from ``indices`` a random index sampled by effective balance.
    """
    assert len(indices) > 0
    MAX_RANDOM_BYTE = 2**8 - 1
    i = uint64(0)
    total = uint64(len(indices))
    while True:
        candidate_index = indices[compute_shuffled_index(i % total, total, seed)]
        random_byte = hash(seed + uint_to_bytes(uint64(i // 32)))[i % 32]
        effective_balance = state.validators[candidate_index].effective_balance
        if effective_balance * MAX_RANDOM_BYTE >= MAX_EFFECTIVE_BALANCE * random_byte:
            return candidate_index
        i += 1
        
        
def compute_committee(indices: Sequence[ValidatorIndex],
                      seed: Bytes32,
                      index: uint64,
                      count: uint64) -> Sequence[ValidatorIndex]:
    """
    Return the committee corresponding to ``indices``, ``seed``, ``index``, and committee ``count``.
    """
    start = (len(indices) * index) // count
    end = (len(indices) * uint64(index + 1)) // count
    return [indices[compute_shuffled_index(uint64(i), uint64(len(indices)), seed)] for i in range(start, end)]
```

