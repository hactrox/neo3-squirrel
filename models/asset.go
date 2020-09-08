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
	Decimals    *big.Float
	Type        string
	TotalSupply *big.Float
}

// Transfer db model.
type Transfer struct {
	BlockIndex uint
	BlockTime  uint64
	TxID       string
	From       string
	To         string
	Amount     *big.Float
}
