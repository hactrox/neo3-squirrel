package address

import (
	"fmt"
	"neo3-squirrel/models"
	"neo3-squirrel/util/log"
	"sync"
)

var (
	// map[address]*models.AddressInfo{}
	addrInfoMap = map[string]*models.AddressInfo{}
	mutex       sync.Mutex
)

// Cache caches address basic info.
func Cache(address string, blockTime uint64) bool {
	mutex.Lock()
	defer mutex.Unlock()

	addrInfo, ok := addrInfoMap[address]
	if !ok {
		addrInfoMap[address] = &models.AddressInfo{
			Address:     address,
			FirstTxTime: blockTime,
			LastTxTime:  blockTime,
			Transfers:   1,
		}

		return true
	}

	if blockTime > addrInfo.LastTxTime {
		addrInfo.LastTxTime = blockTime
	}

	addrInfo.Transfers++

	return false
}

// Init loads all addr info from db to cache.
func Init(array []*models.AddressInfo) {
	mutex.Lock()
	defer mutex.Unlock()

	size := len(addrInfoMap)
	if size > 0 {
		err := fmt.Errorf("address info cache can only be loaded once. Current cache size: %d", size)
		log.Panic(err)
	}

	for _, addrInfo := range array {
		addrInfoMap[addrInfo.Address] = addrInfo
	}
}
