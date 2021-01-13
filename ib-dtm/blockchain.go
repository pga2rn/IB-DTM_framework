package ib_dtm

import (
	"errors"
	"github.com/pga2rn/ib-dtm_framework/config"
	"github.com/pga2rn/ib-dtm_framework/shared/fwtype"
	"github.com/pga2rn/ib-dtm_framework/shared/logutil"
	"sync"
)

// hold blocks
type BeaconBlock struct {
	slot        uint32
	shards      []*ShardBlock
	ptrNext     *BeaconBlock
	ptrPrevious *BeaconBlock
}

type ShardBlock struct {
	skipped        bool
	slot           uint32
	tvoList        map[uint32]*fwtype.TrustValueOffsetsPerSlot // map[slot]sync.map
	proposer       uint32
	votes          []bool // [indexInCommittee]isApproved
	slashing       []uint32
	whistleblowing []uint32
}

type BlockchainRoot struct {
	mu         sync.Mutex
	headSlot   uint32 // the epoch of current head
	headPtr    *BeaconBlock
	blockCount int // total epoch being recorded
	ptrNext    *BeaconBlock
}

func InitBlockchain() *BlockchainRoot {
	return &BlockchainRoot{
		mu:         sync.Mutex{},
		headSlot:   0,
		blockCount: 0,
	}
}

func (head *BlockchainRoot) InitBlockchainBlock(slot uint32, cfg *config.IBDTMConfig) (*BeaconBlock, error) {
	if slot != (head.headSlot+1) && slot != 0 {
		return nil, errors.New("blockchain is out of sync with the simulation")
	}

	// init the new storage object
	storage := &BeaconBlock{
		slot:        slot,
		ptrNext:     nil,
		ptrPrevious: head.headPtr,
		shards:      make([]*ShardBlock, cfg.ShardNum),
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

	for shardId := 0; shardId < cfg.ShardNum; shardId++ {
		storage.shards[shardId] = &ShardBlock{
			skipped:        true,
			slot:           slot,
			tvoList:        make(map[uint32]*fwtype.TrustValueOffsetsPerSlot),
			votes:          make([]bool, cfg.CommitteeSize),
			slashing:       make([]uint32, cfg.SlashingsLimit),
			whistleblowing: make([]uint32, cfg.WhistleBlowingsLimit),
		}
	}

	return storage, nil
}

func (head *BlockchainRoot) GetHeadBlock() *BeaconBlock {
	return head.headPtr
}

func (head *BlockchainRoot) GetBlockForSlot(slot uint32) *BeaconBlock {
	if slot > head.headSlot {
		return nil
	}

	ptr := head.ptrNext
	for i := uint32(0); i < slot; i++ {
		ptr = ptr.ptrNext
	}

	// debug
	if ptr == nil {
		logutil.LoggerList["ib-dtm"].Debugf("[GetBlockForSlot] failed! slot %v", slot)
	}

	return ptr
}
