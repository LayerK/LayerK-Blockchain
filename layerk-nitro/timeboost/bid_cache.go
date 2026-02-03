package timeboost

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type bidCache struct {
	auctionContractDomainSeparator [32]byte
	sync.RWMutex
	bidsByExpressLaneControllerAddr map[common.Address]*ValidatedBid
}

func newBidCache(auctionContractDomainSeparator [32]byte) *bidCache {
	return &bidCache{
		bidsByExpressLaneControllerAddr: make(map[common.Address]*ValidatedBid),
		auctionContractDomainSeparator:  auctionContractDomainSeparator,
	}
}

func (bc *bidCache) add(bid *ValidatedBid) {
	bc.Lock()
	defer bc.Unlock()
	bc.bidsByExpressLaneControllerAddr[bid.ExpressLaneController] = bid
}

// TwoTopBids returns the top two bids for the given chain ID and round
type auctionResult struct {
	firstPlace  *ValidatedBid
	secondPlace *ValidatedBid
}

func (bc *bidCache) size() int {
	bc.RLock()
	defer bc.RUnlock()
	return len(bc.bidsByExpressLaneControllerAddr)

}

type rankedBid struct {
	bid  *ValidatedBid
	hash *big.Int
}

func betterBid(candidate, current *rankedBid) bool {
	if current == nil || current.bid == nil {
		return true
	}
	amountCmp := candidate.bid.Amount.Cmp(current.bid.Amount)
	if amountCmp != 0 {
		return amountCmp > 0
	}
	return candidate.hash.Cmp(current.hash) > 0
}

// topTwoBids returns the top two bids in the cache.
func (bc *bidCache) topTwoBids() *auctionResult {
	bc.RLock()
	defer bc.RUnlock()

	var first *rankedBid
	var second *rankedBid
	for _, bid := range bc.bidsByExpressLaneControllerAddr {
		bidHash := bid.BigIntHash(bc.auctionContractDomainSeparator)
		candidate := &rankedBid{bid: bid, hash: bidHash}

		if betterBid(candidate, first) {
			second = first
			first = candidate
			continue
		}
		if betterBid(candidate, second) {
			second = candidate
		}
	}

	result := &auctionResult{}
	if first != nil {
		result.firstPlace = first.bid
	}
	if second != nil {
		result.secondPlace = second.bid
	}
	return result
}
