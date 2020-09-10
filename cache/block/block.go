package block

import (
	"container/list"
	"neo3-squirrel/models"
	"sync"
)

const maxCachedBlocks = 100

var (
	blockIndexQueue  = list.New()
	lastBlocks       = map[uint]*models.Block{}
	lastTransactions = map[string]*models.Transaction{}

	mutex    sync.Mutex
	bulkMode bool
)

// CacheBlocks caches blocks and transactions in these blocks.
func CacheBlocks(blocks []*models.Block) {
	mutex.Lock()
	bulkMode = true
	defer mutex.Unlock()

	for _, block := range blocks {
		CacheBlock(block)
	}

	bulkMode = false
}

// CacheBlock caches a single block and its transactoins.
func CacheBlock(block *models.Block) {
	if block == nil {
		return
	}

	if !bulkMode {
		mutex.Lock()
		defer mutex.Unlock()
	}

	if _, exists := lastBlocks[block.Index]; exists {
		return
	}

	blockIndexQueue.PushBack(block.Index)
	lastBlocks[block.Index] = block
	for _, tx := range block.GetTxs() {
		lastTransactions[tx.Hash] = tx
	}

	if blockIndexQueue.Len() > maxCachedBlocks {
		firstElem := blockIndexQueue.Front()
		blockIndexToRemove := firstElem.Value.(uint)

		blockIndexQueue.Remove(firstElem)

		blockToRemove, ok := lastBlocks[blockIndexToRemove]
		if !ok {
			return
		}

		for _, tx := range blockToRemove.GetTxs() {
			delete(lastTransactions, tx.Hash)
		}

		delete(lastBlocks, blockIndexToRemove)
	}
}

// GetBlock returns the block of the given index if cached.
func GetBlock(blockIndex uint) (*models.Block, bool) {
	block, ok := lastBlocks[blockIndex]
	return block, ok
}

// GetTransaction returns the transaction if the given hash if cached.
func GetTransaction(txID string) (*models.Transaction, bool) {
	tx, ok := lastTransactions[txID]
	return tx, ok
}
