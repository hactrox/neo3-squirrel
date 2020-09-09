package counter

import (
	"sync/atomic"
)

// SafeCounter defines a thread safe counter.
type SafeCounter struct {
	val int32
}

// Get returns value of the counter.
func (c *SafeCounter) Get() int {
	return int(atomic.LoadInt32(&c.val))
}

// Set sets value of the counter.
func (c *SafeCounter) Set(v int) {
	atomic.StoreInt32(&c.val, int32(v))
}

// Add adds delta to current counter.
func (c *SafeCounter) Add(delta int) int {
	return int(atomic.AddInt32(&c.val, int32(delta)))
}

// SetIfHigher updates current value if the given integer
// is great than the current one.
func (c *SafeCounter) SetIfHigher(v int) bool {
	return atomic.CompareAndSwapInt32(&c.val, int32(v-1), int32(v))
}
