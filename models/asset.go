package models

import (
	"math/big"
)

// NEO & GAS contract hash
const (
	NEO = "0xde5f57d430d3dece511cf975a8d37848cb9e0525"
	GAS = "0x668e0c1f9d7b70a99dd9e06eadd4c784d641afbc"
)

// Asset db model.
type Asset struct {
	ID          uint
	BlockIndex  uint
	BlockTime   uint64
	Contract    string
	Name        string
	Symbol      string
	Decimals    uint
	Type        string
	TotalSupply *big.Float
	Addresses   uint
	Transfers   uint
	// Destroyed   bool
}

// AddrAsset db model.
type AddrAsset struct {
	ID        uint
	Address   string
	Contract  string
	Balance   *big.Float
	Transfers int
}
