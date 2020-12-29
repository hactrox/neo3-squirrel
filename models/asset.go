package models

import (
	"math/big"
)

// NEO & GAS contract hash
const (
	NEO = "0x0a46e2e37c9987f570b4af253fb77e7eef0f72b6"
	GAS = "0xa6a6c15dcdc9b997dac448b6926522d22efeedfb"
)

// Asset db model.
type Asset struct {
	ID          uint
	BlockIndex  uint
	BlockTime   uint64
	TxID        string
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
