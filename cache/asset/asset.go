package asset

import (
	"neo3-squirrel/models"
	"sync"
)

var (
	nep5AssetMap = map[string]*models.Asset{}
	mu           sync.RWMutex
)

// UpdateNEP5Asset adds or updates NEP5 asset cache.
func UpdateNEP5Asset(asset *models.Asset) {
	mu.Lock()
	defer mu.Unlock()

	nep5AssetMap[asset.Contract] = asset
}

// UpdateNEP5Assets adds or updates NEP5 asset caches.
func UpdateNEP5Assets(assets []*models.Asset) {
	mu.Lock()
	defer mu.Unlock()

	for _, asset := range assets {
		nep5AssetMap[asset.Contract] = asset
	}
}

// GetNEP5 returns NEP5 asset from cache if exists.
func GetNEP5(contract string) (*models.Asset, bool) {
	mu.RLock()
	defer mu.RUnlock()

	nep5, ok := nep5AssetMap[contract]
	return nep5, ok
}

// GetNEP5Decimals returns NEP5 asset decimals from cache if exists.
func GetNEP5Decimals(contract string) (uint, bool) {
	mu.RLock()
	defer mu.RUnlock()

	nep5, ok := nep5AssetMap[contract]
	if !ok {
		return 0, false
	}

	return nep5.Decimals, true
}

// GetNEP5Symbol returns NEP5 asset symbol from cache if exists.
func GetNEP5Symbol(contract string) (string, bool) {
	mu.RLock()
	defer mu.RUnlock()

	nep5, ok := nep5AssetMap[contract]
	if !ok {
		return "", false
	}

	return nep5.Symbol, true
}
