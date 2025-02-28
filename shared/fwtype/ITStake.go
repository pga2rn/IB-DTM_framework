package fwtype

import (
	"errors"
	"sync"
)

type ITStake struct {
	mu   sync.RWMutex
	slot []uint32
	m    map[uint32]float32 // map[slot]stake
}

func NewITStack(length int, init float32) *ITStake {
	res := ITStake{
		mu:   sync.RWMutex{},
		slot: make([]uint32, length),
		m:    make(map[uint32]float32),
	}
	// add init amount of ITStake for every RSU
	res.AddAmount(0, init)

	return &res
}

func (s *ITStake) AddAmount(epoch uint32, amount float32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.m[epoch]; ok { // the slot is already in the map
		s.m[epoch] += amount
	} else { // this is the new slot
		if len(s.m) < len(s.slot) {
			s.push(epoch)
		} else { // storage is full, delete old records
			oldEpoch, _ := s.pop()
			delete(s.m, oldEpoch)

			s.push(epoch)
		}
		s.m[epoch] = amount
	}
}

func (s *ITStake) GetAmount() float32 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := float32(0)
	for _, value := range s.m {
		count += value
	}
	return count
}

func (s *ITStake) push(v uint32) {
	s.slot = append(s.slot, v)
}

func (s *ITStake) pop() (uint32, error) {
	l := len(s.slot)
	if l == 0 {
		return 0, errors.New("empty stack")
	}

	res := s.slot[l-1]
	s.slot = s.slot[:l-1]
	return res, nil
}
