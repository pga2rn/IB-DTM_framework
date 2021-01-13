package ib_dtm

import (
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

func InitValidator(vid uint32, initEffectiveStake float32, itsStakeCacheLength int) *Validator {
	res := &Validator{
		mu:                sync.Mutex{},
		uploadMu:          sync.Mutex{},
		Id:                vid,
		effectiveStake:    initEffectiveStake,
		itsStake:          fwtype.NewITStack(itsStakeCacheLength),
		nextSlotForUpload: 0,
	}
	return res
}

func (v *Validator) SetNextSlotForUpload(slot uint32) {
	v.uploadMu.Lock()
	defer v.uploadMu.Unlock()
	v.nextSlotForUpload = slot
}

func (v *Validator) GetNextSlotForUpload() uint32 {
	v.uploadMu.Lock()
	defer v.uploadMu.Unlock()
	return v.nextSlotForUpload
}

func (v *Validator) AddEffectiveStake(amount float32) {
	v.mu.Lock()
	if amount < 0 && math.Abs(float64(amount)) >= float64(v.effectiveStake) {
		v.effectiveStake = 0
	} else {
		if v.effectiveStake+amount > 32 {
			v.effectiveStake = 32
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
