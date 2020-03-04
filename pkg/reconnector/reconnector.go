package reconnector

import (
	"fmt"
	"math"
	"math/rand"
	"neo3-squirrel/log"
	"sync"
	"sync/atomic"
	"time"
)

type locker struct {
	lockerSlots map[string]*uint32

	mutex *sync.Mutex
}

var l *locker

type reconnFunc func() error

func init() {
	l = &locker{
		lockerSlots: make(map[string]*uint32),
		mutex:       &sync.Mutex{},
	}
}

// Reconnect implements auto-reconnect logic for server clients.
func Reconnect(target string, f reconnFunc) {
	lockerAddr := l.getLocker(target)

	// Lock was held by others
	if !atomic.CompareAndSwapUint32(lockerAddr, 0, 1) {
		for {
			time.Sleep(time.Duration(20+rand.Int31n(30)) * time.Millisecond)
			// Lock released(target is connected)
			if atomic.LoadUint32(lockerAddr) == 0 {
				return
			}
		}
	}

	defer atomic.StoreUint32(lockerAddr, 0)

	log.Warnf("%s connection was lost, try reconnecting...", target)
	retryCnt := 0

	for {
		retryCnt++
		if err := f(); err == nil {
			return
		}

		delay := int64(math.Pow(2, float64((retryCnt-1)%4)))

		log.Warnf("[%s][retry=%d] Waiting for %d seconds before retry db connection",
			target, retryCnt, delay)
		time.Sleep(time.Duration(delay) * time.Second)
	}
}

func (l *locker) getLocker(target string) *uint32 {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	_, ok := l.lockerSlots[target]
	if !ok {
		l.lockerSlots[target] = new(uint32)
	}

	return l.lockerSlots[target]
}

func (l *locker) Slots() int {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return len(l.lockerSlots)
}

func (l locker) String() string {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	out := fmt.Sprintf("%d slot used. ", len(l.lockerSlots))
	out += fmt.Sprintf("LockerSlots size: %d, members: [", len(l.lockerSlots))
	for target, slot := range l.lockerSlots {
		out += fmt.Sprintf("%s: %d, ", target, slot)
	}
	if len(l.lockerSlots) > 0 {
		out = out[:len(out)-2] + "]"
	} else {
		out += "]"
	}

	return out
}
