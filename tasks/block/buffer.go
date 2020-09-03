package block

import (
	"neo3-squirrel/rpc"
	"sync"
)

// Buffer is used to temporarily buffer fetched blocks.
type Buffer struct {
	mu sync.Mutex
	// maxHeight indicates the highest existing height.
	maxHeight int
	// nextHeight indicates the next block height to fetch,
	// used before blockchain fully synchronized.
	nextHeight int
	buffer     map[int]*rpc.Block
}

// NewBuffer inits a new block buffer.
func NewBuffer(height int) Buffer {
	return Buffer{
		maxHeight:  height,
		nextHeight: height,
		buffer:     make(map[int]*rpc.Block),
	}
}

// Pop pops the specific block by id.
func (b *Buffer) Pop(index int) (*rpc.Block, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if block, ok := b.buffer[index]; ok {
		delete(b.buffer, index)
		return block, true
	}
	return nil, false
}

// GetHighest returns the highest existing block height.
func (b *Buffer) GetHighest() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.maxHeight
}

// GetNextPending returns the next fetching block index.
func (b *Buffer) GetNextPending() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.nextHeight++
	return b.nextHeight
}

// Put adds the given block into buffer and update maxHeight.
func (b *Buffer) Put(block *rpc.Block) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.buffer[int(block.Index)] = block
	if b.maxHeight < int(block.Index) {
		b.maxHeight = int(block.Index)
	}
}

// Size returns size of current buffer.
func (b *Buffer) Size() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return len(b.buffer)
}
