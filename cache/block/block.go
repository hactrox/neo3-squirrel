package block

import (
	"container/list"
	"neo3-squirrel/models"
	"sync"
)

const maxCachedBlocks = 50000

var (
	blockIndexQueue = list.New()
	blockMap        = map[uint]*models.Block{}
	transactionMap  = map[string]*models.Transaction{}

	mu sync.Mutex

	// bulkMode prevents double mutex lock
	// when caches blocks at once.
	bulkMode bool
)

// CacheBlocks caches blocks and transactions in these blocks.
func CacheBlocks(blocks []*models.Block) {
	mu.Lock()
	defer mu.Unlock()

	bulkMode = true
	defer func() { bulkMode = false }()

	for _, block := range blocks {
		CacheBlock(block)
	}
}

// CacheBlock caches a single block and its transactoins.
// If cached block size exceeded the limite,
// lowerest indexes will be removed.
func CacheBlock(block *models.Block) {
	if block == nil {
		return
	}

	if !bulkMode {
		mu.Lock()
		defer mu.Unlock()
	}

	if _, exists := blockMap[block.Index]; exists {
		return
	}

	blockIndexQueue.PushBack(block.Index)
	blockMap[block.Index] = block
	for _, tx := range block.GetTxs() {
		transactionMap[tx.Hash] = tx
	}

	// Remove oldest cached block if exceed the maximum limitation.
	if blockIndexQueue.Len() > maxCachedBlocks {
		firstElem := blockIndexQueue.Front()
		blockIndexToRemove := firstElem.Value.(uint)

		blockIndexQueue.Remove(firstElem)

		blockToRemove, ok := blockMap[blockIndexToRemove]
		if !ok {
			return
		}

		for _, tx := range blockToRemove.GetTxs() {
			delete(transactionMap, tx.Hash)
		}

		delete(blockMap, blockIndexToRemove)
	}
}

// GetBlock returns the block of the given index if cached.
func GetBlock(blockIndex uint) (*models.Block, bool) {
	mu.Lock()
	defer mu.Unlock()

	block, ok := blockMap[blockIndex]
	return block, ok
}

// GetTransaction returns the transaction if the given hash if cached.
func GetTransaction(txID string) (*models.Transaction, bool) {
	mu.Lock()
	defer mu.Unlock()

	tx, ok := transactionMap[txID]
	return tx, ok
}
