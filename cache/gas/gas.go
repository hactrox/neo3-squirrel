package gas

import (
	"math/big"
	"sync"
)

var (
	blockIndex  uint
	totalSupply *big.Float
	mu          sync.Mutex
)

// CacheGASTotalSupply caches latest GAS total supply.
func CacheGASTotalSupply(newBlockIndex uint, newTotalSupply *big.Float) {
	mu.Lock()
	defer mu.Unlock()

	if newBlockIndex <= blockIndex {
		return
	}

	totalSupply = newTotalSupply
}

// CachedTillBlockIndex returns the block height
// which GAS total supply was queried.
func CachedTillBlockIndex() uint {
	return blockIndex
}

// GetTotalSupply returns cached GAS total supply
// recorded in which block index.
func GetTotalSupply() (*big.Float, uint) {
	mu.Lock()
	defer mu.Unlock()

	return totalSupply, blockIndex
}
