package asset

import (
	"neo3-squirrel/models"
	"sync"
)

var (
	// nep5Assets stores all NEP5 asset info.
	nep5Assets = map[string]*models.Asset{}
	mu         sync.RWMutex
)

// UpdateNEP5Asset adds or updates NEP5 asset cache.
func UpdateNEP5Asset(asset *models.Asset) {
	mu.Lock()
	mu.Unlock()

	nep5Assets[asset.Contract] = asset
}

// UpdateNEP5Assets adds or updates NEP5 asset caches.
func UpdateNEP5Assets(assets []*models.Asset) {
	mu.Lock()
	mu.Unlock()

	for _, asset := range assets {
		nep5Assets[asset.Contract] = asset
	}
}

// GetNEP5 returns NEP5 asset from cache if exists.
func GetNEP5(contract string) (*models.Asset, bool) {
	mu.RLock()
	mu.RUnlock()

	nep5, ok := nep5Assets[contract]
	return nep5, ok
}

// GetNEP5Decimals returns NEP5 asset decimals from cache if exists.
func GetNEP5Decimals(contract string) (uint, bool) {
	mu.RLock()
	mu.RUnlock()

	nep5, ok := nep5Assets[contract]
	if !ok {
		return 0, false
	}

	return nep5.Decimals, true
}

// GetNEP5Symbol returns NEP5 asset symbol from cache if exists.
func GetNEP5Symbol(contract string) (string, bool) {
	mu.RLock()
	mu.RUnlock()

	nep5, ok := nep5Assets[contract]
	if !ok {
		return "", false
	}

	return nep5.Symbol, true
}
