package reconnector

import (
	"errors"
	"neo3-squirrel/util/log"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

func TestGetLocker(t *testing.T) {
	targetA := "Target A"
	lockerAddrA := l.getLocker(targetA)
	if *lockerAddrA != 0 {
		t.Errorf("Failed to get new locker.")
	}
	if l.Slots() != 1 {
		t.Errorf("Locker slot error, get %d, want 1", l.Slots())
	}
	if !atomic.CompareAndSwapUint32(lockerAddrA, 0, 1) {
		t.Errorf("Failed to get locker, this operation should not fail.")
	}

	targetB := "Target B"
	lockerAddrB := l.getLocker(targetB)
	if *lockerAddrB != 0 {
		t.Errorf("Failed to get new locker for target B.")
	}
	if l.Slots() != 2 {
		t.Errorf("Locker slot error, get %d, want 2", l.Slots())
	}
	if !atomic.CompareAndSwapUint32(lockerAddrB, 0, 1) {
		t.Errorf("Failed to get locker, this operation should not fail.")
	}
}

func TestReconnect(t *testing.T) {
	defer func() {
		os.RemoveAll("./logs")
	}()

	log.Init(true)
	connErr := errors.New("error")

	f := func() error {
		defer func() {
			connErr = nil
		}()
		return connErr
	}

	ch := make(chan bool, 1)
	go func() {
		Reconnect("test", f)
		ch <- true
	}()

	select {
	case <-ch:
	case <-time.After(1020 * time.Millisecond):
		t.Fatalf("Failed to reconnect")
	}
}
