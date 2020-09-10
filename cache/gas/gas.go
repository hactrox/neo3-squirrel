package gas

import (
	"math/big"
	"sync"
)

var (
	blockIndex  uint
	totalSupply *big.Float
	mutex       sync.Mutex
)

// CacheGASTotalSupply caches latest GAS total supply.
func CacheGASTotalSupply(newBlockIndex uint, newTotalSupply *big.Float) {
	mutex.Lock()
	defer mutex.Unlock()

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
