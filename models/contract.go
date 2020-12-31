package models

import (
	"encoding/json"
	"neo3-squirrel/rpc"
	"neo3-squirrel/util/log"
)

// Neo3 Native Contracts
const (
	DesignationContract = "0xc0073f4c7069bf38995780c9da065f9b3949ea7a"
	OracleContract      = "0xb1c37d5847c2ae36bdde31d0cc833a7ad9667f8f"
	PolicyContract      = "0xdde31084c0fdbebc7f5ed5f53a38905305ccee14"
	GASContract         = "0xa6a6c15dcdc9b997dac448b6926522d22efeedfb"
	NEOContract         = "0x0a46e2e37c9987f570b4af253fb77e7eef0f72b6"
	ManagementContract  = "0xcd97b70d82d69adfcd9165374109419fade8d6ab"
)

// NativeContractHashes returns all 6 Neo3 native contracts.
func NativeContractHashes() []string {
	return []string{
		DesignationContract,
		OracleContract,
		PolicyContract,
		GASContract,
		NEOContract,
		ManagementContract,
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
func ParseContractState(blockIndex uint, blockTime uint64, txID string, rawCS *rpc.ContractState) *ContractState {
	cs := &ContractState{
		BlockIndex:    blockIndex,
		BlockTime:     blockTime,
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
