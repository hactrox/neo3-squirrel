package contractstate

import (
	"container/list"
	"fmt"
	"log"
	"neo3-squirrel/models"
	"sync"
)

var (
	contractStates = list.New()
	mu             sync.RWMutex
)

// Init loads all contract state from db into cache.
func Init(data [][]*models.ContractState) {
	mu.Lock()
	defer mu.Unlock()

	l := contractStates.Len()
	if l > 0 {
		err := fmt.Errorf("contract state cache can only be loaded once. Current len: %d", l)
		log.Panic(err)
	}

	for _, list := range data {
		if len(list) == 0 {
			continue
		}
		contractStates.PushBack(list)
	}
}

// AddContractState caches contract states,
// all contract states must have the same block index.
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
