package models

import (
	"encoding/json"
	"math/big"
	"neo3-squirrel/rpc"
	"neo3-squirrel/util/log"
)

// ContractState db model.
type ContractState struct {
	ID          uint
	BlockIndex  uint
	BlockTime   uint64
	TxID        string
	Hash        string
	State       string
	NewHash     string
	ContractID  int
	Name        string
	Symbol      string
	Decimals    uint
	TotalSupply *big.Float
	Script      string
	Manifest    []byte
}

// ParseContractState parses struct *rpc.ContractState to *models.ContractState.
func ParseContractState(cs *rpc.ContractState) *ContractState {
	manifest, err := json.Marshal(cs.Manifest)
	if err != nil {
		log.Panic(err)
	}

	return &ContractState{
		BlockIndex:  cs.BlockIndex,
		BlockTime:   cs.BlockTime,
		TxID:        cs.TxID,
		Hash:        cs.Hash,
		State:       cs.State,
		NewHash:     "",
		ContractID:  cs.ID,
		Name:        cs.Name,
		Symbol:      cs.Symbol,
		Decimals:    cs.Decimals,
		TotalSupply: cs.TotalSupply,
		Script:      cs.Script,
		Manifest:    manifest,
	}
}
