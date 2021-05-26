package models

import (
	"encoding/json"
	"neo3-squirrel/rpc"
	"neo3-squirrel/util/log"
)

// Neo3 Native Contracts
const (
	ContractManagement = "0xfffdc93764dbaddd97c48f252a53ea4643faa3fd"
	StdLib             = "0xacce6fd80d44e1796aa0c2c625e9e4e0ce39efc0"
	CryptoLib          = "0x726cb6e0cd8628a1350a611384688911ab75f51b"
	LedgerContract     = "0xda65b600f7124ce6c79950c1772a36403104f2be"
	NeoToken           = "0xef4073a0f2b305a38ec4050e4d3d28bc40ea63f5"
	GasToken           = "0xd2a4cff31913016155e38e474a2c06d08be276cf"
	PolicyContract     = "0xcc5e4edd9f5f8dba8bb65734541df7a1c081c67b"
	RoleManagement     = "0x49cf4e5378ffcd4dec034fd98a174c5491e395e2"
	OracleContract     = "0xfe924b7cfe89ddd271abaf7210a80a7e11178758"
)

// NativeContractHashes returns all 6 Neo3 native contracts.
func NativeContractHashes() []string {
	return []string{
		ContractManagement,
		StdLib,
		CryptoLib,
		LedgerContract,
		NeoToken,
		GasToken,
		PolicyContract,
		RoleManagement,
		OracleContract,
	}
}

// EventName defines notification event name type.
type EventName string

// Contract management events.
const (
	ContractDeployEvent  EventName = "Deploy"
	ContractUpdateEvent  EventName = "Update"
	ContractDestroyEvent EventName = "Destroy"
)

// ContractState db model.
type ContractState struct {
	ID            uint
	BlockIndex    uint
	BlockTime     uint64
	Creator       string
	TxID          string
	ContractID    int
	UpdateCounter uint
	Hash          string
	NEF           NEF
	State         string
	Script        string
	Manifest      ContractManifest
}

type NEF struct {
	Magic    uint64
	Compiler string
	Tokens   []byte
	Script   string
	CheckSum uint64
}

// ContractManifest db model.
type ContractManifest struct {
	Name               string
	Groups             []byte
	Features           []byte
	SupportedStandards []string
	ABI                *ABI
	Permissions        []byte
	Trusts             []byte
	Extra              []byte
}

// ParseContractState parses struct *rpc.ContractState to *models.ContractState.
func ParseContractState(blockIndex uint, blockTime uint64, creator, txID string, rawCS *rpc.ContractState) *ContractState {
	if rawCS == nil {
		return nil
	}

	cs := &ContractState{
		BlockIndex:    blockIndex,
		BlockTime:     blockTime,
		Creator:       creator,
		TxID:          txID,
		ContractID:    rawCS.ID,
		UpdateCounter: rawCS.UpdateCounter,
		Hash:          rawCS.Hash,
		State:         string(ContractDeployEvent), // Set default state
	}

	cs.NEF = NEF{
		Magic:    rawCS.NEF.Magic,
		Compiler: rawCS.NEF.Compiler,
		Tokens:   marshalField(rawCS.NEF.Tokens),
		Script:   rawCS.NEF.Script,
		CheckSum: rawCS.NEF.CheckSum,
	}

	cs.Manifest = ContractManifest{
		Name:               rawCS.Manifest.Name,
		Groups:             marshalField(rawCS.Manifest.Groups),
		Features:           marshalField(rawCS.Manifest.Features),
		SupportedStandards: rawCS.Manifest.SupportedStandards,
		ABI:                unmarshalABI(marshalField(rawCS.Manifest.ABI)),
		Permissions:        marshalField(rawCS.Manifest.Permissions),
		Trusts:             marshalField(rawCS.Manifest.Trusts),
		Extra:              marshalField(rawCS.Manifest.Extra),
	}

	return cs
}

// MarshalSupportedStandards is the shortcut of json.Marshal(cs.Manifest.SupportedStandards).
func (cs *ContractState) MarshalSupportedStandards() []byte {
	return marshalField(cs.Manifest.SupportedStandards)
}

// MarshalABI is the shortcut of json.Marshal(cs.Manifest.ABI).
func (cs *ContractState) MarshalABI() []byte {
	return marshalField(cs.Manifest.ABI)
}

// UnmarshalSupportedStandards is the shortcut of json.Unmarshal(cs.Manifest.SupportedStandards).
func (cs *ContractState) UnmarshalSupportedStandards(supportedStandards []byte) {
	err := json.Unmarshal(supportedStandards, &cs.Manifest.SupportedStandards)
	if err != nil {
		log.Panic(err)
	}
}

func (cs *ContractState) UnmarshalABI(raw []byte) {
	cs.Manifest.ABI = unmarshalABI(raw)
}

func unmarshalABI(raw []byte) *ABI {
	abi := ABI{}
	err := json.Unmarshal(raw, &abi)
	if err != nil {
		log.Panic(err)
	}

	return &abi
}

func marshalField(field interface{}) []byte {
	dat, err := json.Marshal(field)
	if err != nil {
		log.Panic(err)
	}

	return dat
}
