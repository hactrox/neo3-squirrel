package models

import "math/big"

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
}

// Transfer db model.
type Transfer struct {
	ID         uint
	BlockIndex uint
	BlockTime  uint64
	TxID       string
	Contract   string
	From       string
	To         string
	Amount     *big.Float
}

// AddrAsset db model.
type AddrAsset struct {
	ID        uint
	Address   string
	Contract  string
	Balance   *big.Float
	Transfers int
}
