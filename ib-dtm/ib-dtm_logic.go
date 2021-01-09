package ib_dtm

import (
	"context"
	"errors"
	"github.com/boljen/go-bitmap"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/shared/fwtype"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"github.com/pga2rn/ib-dtm_framework/shared/randutil"
	"math"
	"sync"
)

type Validator struct {
	Id uint32
	ITStake float32
	EffectiveStake float32
}

type BeaconStatus struct {
	SimConfig *config.SimConfig
	IBDTMConfig *config.IBDTMConfig

	Epoch            uint32
	activeValidators *map[uint32]*Validator
	validators *[]Validator
	validatorBitMap  *bitmap.Threadsafe
	slashings        *bitmap.Threadsafe
	whistleBlowings  *map[uint32]int

	// committees
	assignment []uint32
	proposer []uint32 // [committeeId]uint32

	// random source
	R *randutil.RandUtil
}

type ValidatorRole = int

const (
	Attestor = iota
	Proposer
)

func (bs *BeaconStatus) IsValidatorActive(vid uint32)bool{
	return bs.validatorBitMap.Get(int(vid))
}

func (bs *BeaconStatus) InactivateValidator(vid uint32){
	delete(*bs.activeValidators, vid)
	bs.validatorBitMap.Set(int(vid), false)
}

func (bs *BeaconStatus) ActivateValidator(vid uint32){
	(*bs.activeValidators)[vid] = &(*bs.validators)[int(vid)]
	bs.validatorBitMap.Set(int(vid), true)
}

// get rid of the validators that with low balance
func (bs *BeaconStatus) ProcessValidatorLifeCycle(ctx context.Context, epoch uint32){
	if epoch != bs.Epoch{
		logutil.LoggerList["ib-dtm"].Fatalf("[ProcessValidatorLifeCycle] failed")
	}

	select {
	case <-ctx.Done():
		logutil.LoggerList["ib-dtm"].Fatalf("[ProcessValidatorLifeCycle] context canceled")
	default:
		// process slashings
		for i := 0;i<bs.SimConfig.RSUNum;i++{
			if bs.slashings.Get(i){
				delete(*bs.ActiveValidators, uint32(i))
			}
			if bs.ActiveValidators[i].EffectiveStake < bs.SimConfig
		}
	}
}

func (bs *BeaconStatus) PrepareForNextEpoch(epoch uint32){
	// clear the storage area
	bs.slashings = bitmap.NewTS(bs.SimConfig.RSUNum)
	bs.whistleBlowings = new(map[uint32]int)

	// regenerate the committee
	bs.committeeSize = len(*bs.ActiveValidators)
}

// generate assignment
func (bs *BeaconStatus) genAssignment(){
	// generate validator id list
	vidList := make([]uint32, len(*bs.ActiveValidators))
	bs.assignment = make([]uint32, len(*bs.ActiveValidators))

	count := -1
	for vid, _ := range *bs.ActiveValidators {
		count ++
		vidList[count] = vid
	}

	// shuffle and gen assignment
	shuffledIndex := bs.R.Perm(count)
	for i, index := range shuffledIndex{
		bs.assignment[i] = vidList[index]
	}

	// gen proposer
	for i := 0; i < int(bs.committeeNum); i++{
		committee := bs.assignment[i:i+int(bs.committeeSize)]
		bs.proposer[i] = committee[bs.R.Intn(int(bs.committeeSize))]
	}
}

// get the committee for rid
func (bs *BeaconStatus) GetCommitteeId(rid uint32) (uint32, error) {
	for i, index := range bs.assignment{
		if index == rid{
			return uint32(i) / bs., nil
		}
	}
	return 65536, errors.New("failed to get committeeId")
}

func (bs *BeaconStatus) GetRole(rid uint32) (ValidatorRole, error) {
	committee, err := bs.GetCommitteeId(rid)
	if err != nil{
		return 65536, err
	}

	if bs.proposer[committee] == rid{
		return Proposer, nil
	} else {
		return Attestor, nil
	}
}

func (session *IBDTMSession) calculateTrustValueHelper(
	tvo *fwtype.TrustValueOffset,
	compromisedRSUFlag bool) float32 {

	res := tvo.Weight * tvo.TrustValueOffset / float32(session.SimConfig.SlotsPerEpoch)

	if compromisedRSUFlag {
		switch tvo.AlterType {
		case fwtype.Flipped:
			res = -res
		case fwtype.Dropped:
			res = 0
		}
	}
	return res
}


func (session *IBDTMSession) genTrustValue(ctx context.Context, epoch uint32) {
	// iterate through the blockchain for all experiments
	for _, exp := range session.ExpConfigList{
		// init storage area
		session.TrustValueStorage[exp.Name] = &fwtype.TrustValuesPerEpoch{}
		blockchain := session.Blockchain[exp.Name]
		result := session.TrustValueStorage[exp.Name]

		startSlot, endSlot := uint32(0), (epoch+1) * session.SimConfig.SlotsPerEpoch
		if epoch < uint32(exp.TrustValueOffsetsTraceBackEpochs){
			startSlot = 0
		} else {
			startSlot = endSlot - uint32(exp.TrustValueOffsetsTraceBackEpochs) * session.SimConfig.SlotsPerEpoch
		}

		// iterate through each slots
		wg := sync.WaitGroup{}
		// for each shard
		for i := startSlot; i < endSlot; i++ {
			block := blockchain.GetBlockForSlot(i)

			wg.Add(1)
			go func() {
				// for each shard
				for _, shard := range block.shards {
					// dive into the slot
					c := make(chan []interface{})
					// define a call back function to take the value out of sync.map
					f := func(key, value interface{}) bool {
						c <- []interface{}{key, value}
						return true
					}

					// capture all values in the slot
					go func(){
						for pair := range c{
							key, value := pair[0].(uint32), pair[1].(*fwtype.TrustValueOffset)
							if key != value.VehicleId {
								logutil.LoggerList["simulator"].
									Warnf("[genBaselineTrustValue] mismatch vid! %v in vehicle and %v in tvo", key, value.VehicleId)
								continue // ignore invalid trust value offset record
							}

							tvo := session.calculateTrustValueHelper(value, exp.CompromisedRSUFlag)
							if op, ok := result.LoadOrStore(value.VehicleId, tvo); ok{
								result.Store(value.VehicleId, tvo+op.(float32))
							}
						}
					}()

					shard.tvoList.Range(f)
					close(c)
				}// iterate shards
			}()// iterate slots
			wg.Done()
		}

		wg.Wait()
	}
}