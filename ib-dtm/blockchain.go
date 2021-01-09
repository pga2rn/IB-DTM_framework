package ib_dtm

import (
	"errors"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/shared/fwtype"
	"sync"
)

// hold blocks
type BlockchainBlock struct {
	slot        uint32
	shards      []*BlockchainShard
	ptrNext     *BlockchainBlock
	ptrPrevious *BlockchainBlock
}

type BlockchainShard struct {
	skipped        bool
	slot           uint32
	tvoList        *fwtype.TrustValueOffsetsPerSlot
	proposer       uint32
	votes          *map[uint32]bool
	slashing       []uint32
	whistleblowing []uint32
}

type BlockchainHead struct {
	mu         sync.Mutex
	headSlot   uint32 // the epoch of current head
	headPtr    *BlockchainBlock
	blockCount int // total epoch being recorded
	ptrNext    *BlockchainBlock
}

func InitBlockchain() *BlockchainHead {
	return &BlockchainHead{
		mu:         sync.Mutex{},
		headSlot:   0,
		blockCount: 0,
	}
}

// TODO: other parts of blockchain block!
func (head *BlockchainHead) InitBlockchainBlock(slot uint32, cfg *config.SimConfig) (*BlockchainBlock, error) {
	if slot != (head.headSlot+1) && slot != 0 {
		return nil, errors.New("blockchain is out of sync with the simulation")
	}

	// init the new storage object
	storage := &BlockchainBlock{
		slot:        slot,
		ptrNext:     nil,
		ptrPrevious: head.headPtr,
	}

	head.mu.Lock()
	// update the head block
	if head.headPtr != nil {
		head.headPtr.ptrNext = storage
		head.headPtr = storage
	} else {
		// for slot 0
		head.ptrNext = storage
		head.headPtr = storage
	}

	// update the head information in the head
	head.headSlot, head.blockCount = slot, head.blockCount+1
	head.mu.Unlock()

	return storage, nil
}

func (head *BlockchainHead) GetHeadBlock() *BlockchainBlock {
	return head.headPtr
}

func (head *BlockchainHead) GetBlockForSlot(slot uint32) *BlockchainBlock {
	if slot > head.headSlot {
		return nil
	}

	ptr := head.ptrNext
	for i := uint32(0); i < slot; i++ {
		ptr = ptr.ptrNext
	}
	return ptr
}
