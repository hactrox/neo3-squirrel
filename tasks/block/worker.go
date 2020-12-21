package block

import (
	"sync"
)

// Worker shows how many goroutines currently running for block data persistence.
type Worker struct {
	mu           sync.Mutex
	goroutineCnt uint8
}

func (manager *Worker) shouldQuit() bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if manager.goroutineCnt > 1 {
		manager.goroutineCnt--
		return true
	}
	return false
}

func (manager *Worker) num() uint8 {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.goroutineCnt
}

func (manager *Worker) add() uint8 {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.goroutineCnt++

	return manager.goroutineCnt
}
