package contractstate

import (
	"container/list"
	"log"
	"neo3-squirrel/models"
	"reflect"
	"sync"
)

var (
	contractStates = list.New()
	mu             sync.RWMutex
)

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
	list, ok := firstElem.Value.([]*models.ContractState)
	if !ok {
		log.Panicf("Failed to extract contract state from cache, get data type=%s",
			reflect.TypeOf(firstElem.Value).String())
	}

	if len(list) == 0 {
		return nil
	}

	if blockIndex <= list[0].BlockIndex {
		return nil
	}

	contractStates.Remove(firstElem)
	return list
}
