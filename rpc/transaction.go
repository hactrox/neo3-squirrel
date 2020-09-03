package rpc

import "math/big"

// Signer is the raw transaction signer structure.
type Signer struct {
	Account string `json:"account"`
	Scopes  string `json:"scopes"`
}

// TxAttribute is the raw transaction attribute structure.
type TxAttribute struct {
	Type string `json:"type"`
}

// Tx is the transaction part of block data.
type Tx struct {
	Hash            string        `json:"hash"`
	Size            uint          `json:"size"`
	Version         uint          `json:"version"`
	Nonce           uint64        `json:"nonce"`
	Sender          string        `json:"sender"`
	SysFee          *big.Float    `json:"sysfee"`
	NetFee          *big.Float    `json:"netfee"`
	ValidUntilBlock int           `json:"validuntilblock"`
	Signers         []Signer      `json:"signers"`
	Attributes      []TxAttribute `json:"attributes"`
	Script          string        `json:"script"`
	Witnesses       []Witness     `json:"witnesses"`
}
