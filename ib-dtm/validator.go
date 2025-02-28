package ib_dtm

import (
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/shared/fwtype"
	"math"
	"sync"
)

type Validator struct {
	mu             sync.Mutex
	Id             uint32
	effectiveStake float32

	itsStake *fwtype.ITStake

	nextSlotForUpload uint32 // the slot that available for uploading trust value offset
	uploadMu          sync.Mutex
}

func InitValidator(vid uint32, cfg *config.IBDTMConfig, exp *config.ExperimentConfig) *Validator {
	res := &Validator{
		mu:                sync.Mutex{},
		uploadMu:          sync.Mutex{},
		Id:                vid,
		effectiveStake:    cfg.InitialEffectiveStake,
		itsStake:          fwtype.NewITStack(exp.TrustValueOffsetsTraceBackEpochs, cfg.InitialITStake),
		nextSlotForUpload: 0,
	}
	return res
}

func (v *Validator) SetNextSlotForUpload(slot uint32) {
	v.uploadMu.Lock()
	defer v.uploadMu.Unlock()

	if slot < v.nextSlotForUpload {
		return
	}
	v.nextSlotForUpload = slot
}

func (v *Validator) GetNextSlotForUpload() uint32 {
	v.uploadMu.Lock()
	defer v.uploadMu.Unlock()
	return v.nextSlotForUpload
}

func (v *Validator) AddEffectiveStake(amount float32, cfg *config.IBDTMConfig) {
	v.mu.Lock()
	if amount < 0 && math.Abs(float64(amount)) >= float64(v.effectiveStake) {
		v.effectiveStake = 0
	} else {
		if v.effectiveStake+amount > cfg.EffectiveStakeUpperBound {
			v.effectiveStake = cfg.EffectiveStakeUpperBound
		} else {
			v.effectiveStake += amount
		}
	}
	v.mu.Unlock()
}

func (v *Validator) AddITStake(epoch uint32, amount float32) {
	v.itsStake.AddAmount(epoch, amount)
}

func (v *Validator) GetITStake() float32 {
	return v.itsStake.GetAmount()
}

// for sorting the validator with effective stake
type Validators []*Validator
type ValidatorPointerList struct {
	Validators
}

func (v Validators) Len() int {
	return len(v)
}

func (v Validators) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

// descend
func (vl ValidatorPointerList) Less(i, j int) bool {
	return vl.Validators[i].effectiveStake > vl.Validators[j].effectiveStake
}
