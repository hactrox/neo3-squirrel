package contractstate

import (
	"container/list"
	"neo3-squirrel/models"
	"sync"
)

var (
	contractStates = list.New()
	mu             sync.RWMutex
)

// AddContractState caches contract states.
func AddContractState(list []*models.ContractState) {
	mu.Lock()
	defer mu.Unlock()

	contractStates.PushBack(list)
}

// PopFirstIf returns the first contract state if its
// block index is lower than the given block index.
func PopFirstIf(blockIndex uint) []*models.ContractState {
	mu.RLock()
	defer mu.RUnlock()

	if contractStates.Len() == 0 {
		return nil
	}

	firstElem := contractStates.Front()
	list := firstElem.Value.([]*models.ContractState)
	if len(list) == 0 {
		return nil
	}

	if list[0].BlockIndex >= blockIndex {
		return nil
	}

	contractStates.Remove(firstElem)
	return list
}
