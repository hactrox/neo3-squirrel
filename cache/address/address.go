package address

import (
	"fmt"
	"neo3-squirrel/util/log"
	"sync"
)

var (
	// map[address]bool
	addrMap = map[string]bool{}
	mutex   sync.Mutex
)

// Cache caches the given address.
// Returns `true` if the address has not been cached yet.
func Cache(address string) bool {
	mutex.Lock()
	defer mutex.Unlock()

	_, ok := addrMap[address]
	if !ok {
		addrMap[address] = true
		return true
	}

	return false
}

// Init loads all addresses from db to cache.
func Init(addresses []string) {
	mutex.Lock()
	defer mutex.Unlock()

	size := len(addrMap)
	if size > 0 {
		err := fmt.Errorf("address cache can only be loaded once. Current cache size: %d", size)
		log.Panic(err)
	}

	for _, addr := range addresses {
		addrMap[addr] = true
	}
}
