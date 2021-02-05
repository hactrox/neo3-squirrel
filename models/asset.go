package models

import (
	"math/big"
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
	TotalSupply *big.Float
	Addresses   uint
	Transfers   uint
}

// AddrAsset db model.
type AddrAsset struct {
	ID        uint
	Address   string
	Contract  string
	Balance   *big.Float
	Transfers int
}
