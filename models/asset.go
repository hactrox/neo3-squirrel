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

// IsGASClaimTransfer tells if this transfer is GAS claim transfer.
func (transfer *Transfer) IsGASClaimTransfer() bool {
	if transfer.Contract == GAS &&
		transfer.From == "" &&
		transfer.To != "" {
		return true
	}

	return false
}
