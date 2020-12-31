package asset

import (
	"neo3-squirrel/models"
	"sync"
)

var (
	assetMap = map[string]*models.Asset{}
	mu       sync.RWMutex
)

// Update adds or updates asset cache.
func Update(asset *models.Asset) {
	mu.Lock()
	defer mu.Unlock()

	assetMap[asset.Contract] = asset
}

// UpdateMulti adds or updates multiple asset caches.
func UpdateMulti(assets []*models.Asset) {
	mu.Lock()
	defer mu.Unlock()

	for _, asset := range assets {
		assetMap[asset.Contract] = asset
	}
}

// Get returns asset from cache (if exists).
func Get(hash string) (*models.Asset, bool) {
	mu.RLock()
	defer mu.RUnlock()

	asset, ok := assetMap[hash]
	return asset, ok
}

// Remove removes the given asset from cache.
func Remove(hash string) {
	mu.RLock()
	defer mu.RUnlock()

	delete(assetMap, hash)
}

// GetDecimals returns asset decimals from cache if exists.
func GetDecimals(hash string) (uint, bool) {
	mu.RLock()
	defer mu.RUnlock()

	asset, ok := assetMap[hash]
	if !ok {
		return 0, false
	}

	return asset.Decimals, true
}

// GetSymbol returns asset symbol from cache (if exists).
func GetSymbol(hash string) (string, bool) {
	mu.RLock()
	defer mu.RUnlock()

	asset, ok := assetMap[hash]
	if !ok {
		return "", false
	}

	return asset.Symbol, true
}
