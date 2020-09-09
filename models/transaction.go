package models

import (
	"math/big"
	"neo3-squirrel/rpc"
)

// TxBulk contains splited tx structure for bulk persistant.
type TxBulk struct {
	Txs         []*Transaction
	TxSigners   []*TransactionSigner
	TxAttrs     []*TransactionAttribute
	TxWitnesses []*TransactionWitness
}

// Transaction db model.
type Transaction struct {
	ID              uint
	BlockIndex      uint
	BlockTime       uint64
	Hash            string
	Size            uint
	Version         uint
	Nonce           uint64
	Sender          string
	SysFee          *big.Float
	NetFee          *big.Float
	ValidUntilBlock int
	// Signer array
	// Attribute array
	Script string
	// Witness array
}

// TransactionSigner represents signer structure of tx.
type TransactionSigner struct {
	TransactionHash string
	Account         string
	Scopes          string
}

// TransactionAttribute represents attribute structure of tx.
type TransactionAttribute struct {
	TransactionHash string
	Type            string
}

// TransactionWitness represents witness structure of tx.
type TransactionWitness struct {
	TransactionHash string
	Invocation      string
	Verification    string
}

// ParseTx parses all *rpc.Transaction in the given block to *models.Transaction.
func ParseTx(block *rpc.Block) []*Transaction {
	txs := []*Transaction{}
	for _, tx := range block.Tx {
		txs = appendTx(txs, block.Index, block.Time, &tx)
	}

	return txs
}

// ParseTxs parses all raw transactions in raw blocks to Bulk.
func ParseTxs(blocks []*rpc.Block) *TxBulk {
	bulk := TxBulk{}

	for _, block := range blocks {
		for _, tx := range block.Tx {
			bulk.Txs = appendTx(bulk.Txs, block.Index, block.Time, &tx)
			bulk.TxSigners = appendTxSigners(bulk.TxSigners, &tx)
			bulk.TxAttrs = appendTxAttrs(bulk.TxAttrs, &tx)
			bulk.TxWitnesses = appendTxWitnesses(bulk.TxWitnesses, &tx)
		}
	}

	return &bulk
}

func appendTx(txs []*Transaction, blockIndex uint, blockTime uint64, rawTx *rpc.Tx) []*Transaction {
	tx := Transaction{
		BlockIndex:      blockIndex,
		BlockTime:       blockTime,
		Hash:            rawTx.Hash,
		Size:            rawTx.Size,
		Version:         rawTx.Version,
		Nonce:           rawTx.Nonce,
		Sender:          rawTx.Sender,
		SysFee:          rawTx.SysFee,
		NetFee:          rawTx.NetFee,
		ValidUntilBlock: rawTx.ValidUntilBlock,
		Script:          rawTx.Script,
	}

	txs = append(txs, &tx)
	return txs
}

func appendTxSigners(signers []*TransactionSigner, rawTx *rpc.Tx) []*TransactionSigner {
	for _, rawSigner := range rawTx.Signers {
		signer := TransactionSigner{
			TransactionHash: rawTx.Hash,
			Account:         rawSigner.Scopes,
			Scopes:          rawSigner.Scopes,
		}

		signers = append(signers, &signer)
	}

	return signers
}

func appendTxAttrs(attrs []*TransactionAttribute, rawTx *rpc.Tx) []*TransactionAttribute {
	for _, rawAttr := range rawTx.Attributes {
		attr := TransactionAttribute{
			TransactionHash: rawTx.Hash,
			Type:            rawAttr.Type,
		}

		attrs = append(attrs, &attr)
	}

	return attrs
}

func appendTxWitnesses(witnesses []*TransactionWitness, rawTx *rpc.Tx) []*TransactionWitness {
	for _, rawWitness := range rawTx.Witnesses {
		witness := TransactionWitness{
			TransactionHash: rawTx.Hash,
			Invocation:      rawWitness.Invocation,
			Verification:    rawWitness.Verification,
		}

		witnesses = append(witnesses, &witness)
	}

	return witnesses
}
