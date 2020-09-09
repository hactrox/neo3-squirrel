package util

import (
	"container/list"
	"neo3-squirrel/models"
)

const cachedBlockCount = 100

var (
	blockIndexQueue  = list.New()
	lastBlocks       = map[uint]*models.Block{}
	lastTransactions = map[string]*models.Transaction{}
)

func CacheBlocks(blocks []*models.Block) {
	for _, block := range blocks {
		CacheBlcok(block)
	}
}

func CacheBlcok(block *models.Block) {
	if block == nil {
		return
	}

	if _, exists := lastBlocks[block.Index]; exists {
		return
	}

	blockIndexQueue.PushBack(block.Index)
	lastBlocks[block.Index] = block
	for _, tx := range block.GetTxs() {
		lastTransactions[tx.Hash] = tx
	}

	if blockIndexQueue.Len() > cachedBlockCount {
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

func GetCachedBlock(blockIndex uint) (*models.Block, bool) {
	block, ok := lastBlocks[blockIndex]
	return block, ok
}

func GetCachedTransaction(txID string) (*models.Transaction, bool) {
	tx, ok := lastTransactions[txID]
	return tx, ok
}
