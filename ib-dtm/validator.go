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
}

func InitValidator(vid uint32, initEffectiveStake float32, itsStakeCacheLength int) *Validator {
	res := &Validator{
		mu:             sync.Mutex{},
		Id:             vid,
		effectiveStake: initEffectiveStake,
		itsStake:       fwtype.NewITStack(itsStakeCacheLength),
	}
	return res
}

func (v *Validator) AddEffectiveStake(amount float32) {
	v.mu.Lock()
	if amount < 0 && math.Abs(float64(amount)) >= float64(v.effectiveStake) {
		v.effectiveStake = 0
	} else {
		v.effectiveStake += amount
	}
	v.mu.Unlock()
}

func (v *Validator) AddITStake(epoch uint32, amount float32) {
	v.itsStake.AddAmount(epoch, amount)
}

func (v *Validator) GetITStake() float32 {
	return v.itsStake.GetAmount()
}
