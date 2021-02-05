package models

import (
	"encoding/json"
	"neo3-squirrel/rpc"
	"neo3-squirrel/util/log"
)

// Neo3 Native Contracts
const (
	ManagementContract = "0xa501d7d7d10983673b61b7a2d3a813b36f9f0e43"
	LedgerContract     = "0x971d69c6dd10ce88e7dfffec1dc603c6125a8764"
	NEOContract        = "0xf61eebf573ea36593fd43aa150c055ad7906ab83"
	GASContract        = "0x70e2301955bf1e74cbb31d18c2f96972abadb328"
	PolicyContract     = "0x79bcd398505eb779df6e67e4be6c14cded08e2f2"
	RoleManagement     = "0x597b1471bbce497b7809e2c8f10db67050008b02"
	OracleContract     = "0x8dc0e742cbdfdeda51ff8a8b78d46829144c80ee"
	NameContract       = "0xa2b524b68dfe43a9d56af84f443c6b9843b8028c"
)

// NativeContractHashes returns all 6 Neo3 native contracts.
func NativeContractHashes() []string {
	return []string{
		ManagementContract,
		LedgerContract,
		NEOContract,
		GASContract,
		PolicyContract,
		RoleManagement,
		OracleContract,
		NameContract,
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
