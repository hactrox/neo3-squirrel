package models

import "math/big"

// Transfer db model.
type Transfer struct {
	ID         uint
	BlockIndex uint
	BlockTime  uint64
	TxID       string
	Src        string
	Contract   string
	From       string
	To         string
	Amount     *big.Float
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
