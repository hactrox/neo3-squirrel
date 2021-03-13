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
	NameService        = "0x7a8fcf0392cd625647907afa8e45cc66872b596b"
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
		NameService,
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
	Hash          string
	State         string
	UpdateCounter uint
	Script        string
	Manifest      ContractManifest
}

// ContractManifest db model.
type ContractManifest struct {
	Name               string
	Groups             []byte
	SupportedStandards []string
	ABI                []byte
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
		Hash:          rawCS.Hash,
		State:         string(ContractDeployEvent), // Set default state
		UpdateCounter: rawCS.UpdateCounter,
		Script:        rawCS.Script,
	}

	cs.Manifest = ContractManifest{
		Name:               rawCS.Manifest.Name,
		Groups:             marshalManifestField(rawCS.Manifest.Groups),
		SupportedStandards: rawCS.Manifest.SupportedStandards,
		ABI:                marshalManifestField(rawCS.Manifest.ABI),
		Permissions:        marshalManifestField(rawCS.Manifest.Permissions),
		Trusts:             marshalManifestField(rawCS.Manifest.Trusts),
		Extra:              marshalManifestField(rawCS.Manifest.Extra),
	}

	return cs
}

// MarshalSupportedStandards is the shortcut of json.Marshal(cs.SupportedStandards).
func (cs *ContractState) MarshalSupportedStandards() []byte {
	return marshalManifestField(cs.Manifest.SupportedStandards)
}

// UnMarshalSupportedStandards is the shortcut of json.Unmarshal(cs.SupportedStandards).
func (cs *ContractState) UnMarshalSupportedStandards(supportedStandards []byte) {
	err := json.Unmarshal(supportedStandards, &cs.Manifest.SupportedStandards)
	if err != nil {
		log.Panic(err)
	}
}

func marshalManifestField(field interface{}) []byte {
	dat, err := json.Marshal(field)
	if err != nil {
		log.Panic(err)
	}

	return dat
}
