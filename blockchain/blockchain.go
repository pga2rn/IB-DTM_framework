package blockchain

import (
	"errors"
	"github.com/pga2rn/ib-dtm_framework/config"
	"sync"
)

// hold blocks
type Blockchain struct {
	slot        uint32
	ptrNext     *Blockchain
	ptrPrevious *Blockchain
}

type BlockchainHead struct {
	mu         sync.Mutex
	headSlot   uint32 `the epoch of current head`
	headPtr    *Blockchain
	blockCount int `total epoch being recorded`
	ptrNext    *Blockchain
}

func InitBlockchain() *BlockchainHead {
	return &BlockchainHead{
		mu:         sync.Mutex{},
		headSlot:   0,
		blockCount: 0,
	}
}

// TODO: other parts of blockchain block!
func (head *BlockchainHead) InitBlockchainBlock(slot uint32, cfg *config.SimConfig) (*Blockchain, error) {
	if slot != (head.headSlot+1) && slot != 0 {
		return nil, errors.New("blockchain is out of sync with the simulation")
	}

	// init the new storage object
	storage := &Blockchain{
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

func (head *BlockchainHead) GetHeadBlock() *Blockchain {
	return head.headPtr
}

func (head *BlockchainHead) GetBlockForSlot(slot uint32) *Blockchain {
	if slot > head.headSlot {
		return nil
	}

	ptr := head.ptrNext
	for i := uint32(0); i < slot; i++ {
		ptr = ptr.ptrNext
	}
	return ptr
}
