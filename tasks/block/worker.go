package block

import (
	"sync"
)

// Worker shows how many goroutines currently running for block data persistence.
type Worker struct {
	mu        sync.Mutex
	threadCnt uint8
}

func (manager *Worker) shouldQuit() bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if manager.threadCnt > 1 {
		manager.threadCnt--
		return true
	}
	return false
}

func (manager *Worker) num() uint8 {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.threadCnt
}

func (manager *Worker) add() uint8 {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.threadCnt++

	return manager.threadCnt
}
