package models

import "neo3-squirrel/rpc"

// Block db model.
type Block struct {
	ID                   uint
	Hash                 string
	Size                 int
	Version              uint
	PreviousBlockHash    string
	MerkleRoot           string
	Txs                  uint
	Time                 uint64
	Index                uint
	NextConsensus        string
	Witnesses            []Witness
	ConsensusDataPrimary uint64
	ConsensusDataNonce   string
	NextBlockHash        string

	txs []*Transaction
}

// GetTxs returns all transactions in this block.
func (block *Block) GetTxs() []*Transaction {
	return block.txs
}

// SetTxs sets transactions for the current block.
func (block *Block) SetTxs(txs []*Transaction) {
	block.txs = txs
}

// ParseBlocks parses struct RawBlock to struct Block.
func ParseBlocks(rawBlocks []*rpc.Block) []*Block {
	blocks := []*Block{}

	for _, rawBlock := range rawBlocks {
		block := Block{
			Hash:                 rawBlock.Hash,
			Size:                 rawBlock.Size,
			Version:              rawBlock.Version,
			PreviousBlockHash:    rawBlock.PreviousBlockHash,
			MerkleRoot:           rawBlock.MerkleRoot,
			Txs:                  uint(len(rawBlock.Tx)),
			Time:                 rawBlock.Time,
			Index:                rawBlock.Index,
			NextConsensus:        rawBlock.NextConsensus,
			ConsensusDataNonce:   rawBlock.ConsensusData.Nonce,
			ConsensusDataPrimary: rawBlock.ConsensusData.Primary,
			NextBlockHash:        rawBlock.NextBlockHash,
			txs:                  ParseTx(rawBlock),
		}

		for _, witness := range rawBlock.Witnesses {
			block.Witnesses = append(block.Witnesses, Witness{
				Invocation:   witness.Invocation,
				Verification: witness.Verification,
			})
		}

		blocks = append(blocks, &block)
	}
	return blocks
}
