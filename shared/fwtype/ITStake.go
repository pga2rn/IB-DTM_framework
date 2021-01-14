package fwtype

import (
	"errors"
	"sync"
)

type ITStake struct {
	mu   sync.Mutex
	slot []uint32
	m    map[uint32]float32
}

func NewITStack(length int) *ITStake {
	return &ITStake{
		mu:   sync.Mutex{},
		slot: make([]uint32, length),
		m:    make(map[uint32]float32),
	}
}

func (s *ITStake) AddAmount(epoch uint32, amount float32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.m[epoch]; ok { // the slot is already in the map
		s.m[epoch] += amount
	} else { // this is the new slot
		if len(s.m) < len(s.slot) {
			s.push(epoch)
		} else {
			oldEpoch, _ := s.pop()
			delete(s.m, oldEpoch)

			s.push(epoch)
		}
		s.m[epoch] = amount
	}
}

func (s *ITStake) GetAmount() float32 {
	s.mu.Lock()
	defer s.mu.Unlock()

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
