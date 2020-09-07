package models

import "math/big"

type Transfer struct {
	BlockIndex uint
	BlockTime  uint64
	TxID       string
	From       string
	To         string
	Amount     *big.Float
}
