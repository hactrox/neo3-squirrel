package models

import (
	"encoding/json"
	"math/big"
	"neo3-squirrel/rpc"
	"neo3-squirrel/util/convert"
	"neo3-squirrel/util/log"
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
	Body            []byte
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
		SysFee:          convert.AmountReadable(rawTx.SysFee, 8),
		NetFee:          convert.AmountReadable(rawTx.NetFee, 8),
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
			Account:         rawSigner.Account,
			Scopes:          rawSigner.Scopes,
		}

		signers = append(signers, &signer)
	}

	return signers
}

func appendTxAttrs(attrs []*TransactionAttribute, rawTx *rpc.Tx) []*TransactionAttribute {
	txAttrsBytes := marshalTxAttributes(rawTx.Attributes)
	if string(txAttrsBytes) == "[]" {
		return attrs
	}

	attr := TransactionAttribute{
		TransactionHash: rawTx.Hash,
		Body:            txAttrsBytes,
	}

	attrs = append(attrs, &attr)
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

func marshalTxAttributes(txAttrs interface{}) []byte {
	dat, err := json.Marshal(txAttrs)
	if err != nil {
		log.Panic(err)
	}

	return dat
}
