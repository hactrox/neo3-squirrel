package nep5

import (
	"encoding/json"
	"neo3-squirrel/cache/asset"
	"neo3-squirrel/cache/contractstate"
	"neo3-squirrel/db"
	"neo3-squirrel/models"
	"neo3-squirrel/util/log"
	"strings"
)

// Contract track state string literal.
const (
	Added   = "Added"
	Updated = "Updated"
	Deleted = "Deleted"
)

func handleContractStateChange(minBlockIndex uint) {
	csList := contractstate.PopFirstIf(minBlockIndex)
	if len(csList) == 0 {
		return
	}

	contractStates := []*models.ContractState{}
	added := []*models.ContractState{}
	deleted := []*models.ContractState{}
	migrated := map[*models.ContractState]*models.ContractState{} // map[new]old

	for i := 0; i < len(csList); i++ {
		cs := csList[i]
		// log.Debugf(cs.Hash)

		contractStates = append(contractStates, cs)

		switch cs.State {
		case Added:
			updateIfNEP5(cs)
			added = append(added, cs)
		case Updated:
			// Passed.
		case Deleted:
			// If contract migration.
			if i-1 >= 0 && csList[i-1].TxID == cs.TxID && csList[i-1].State == Added {
				newContract := csList[i-1]
				oldContract := cs
				migrated[newContract] = oldContract
				continue
			}

			deleted = append(deleted, cs)
		default:
			log.Panicf("Unsupported contract track state: %s", cs.State)
		}
	}

	db.HandleContractStates(contractStates, added, deleted, migrated)
}

func updateIfNEP5(c *models.ContractState) {
	if c.Name != "" && c.Symbol != "" {
		manifest, err := json.Marshal(c.Manifest)
		if err != nil {
			log.Panic(err)
		}

		manifestStr := strings.ToLower(string(manifest))
		if strings.Contains(manifestStr, "nep-5") ||
			strings.Contains(manifestStr, "nep5") {
			nep5 := &models.Asset{
				BlockIndex:  c.BlockIndex,
				BlockTime:   c.BlockTime,
				Contract:    c.Hash,
				Name:        c.Name,
				Symbol:      c.Symbol,
				Decimals:    c.Decimals,
				Type:        "nep5",
				TotalSupply: c.TotalSupply,
			}

			asset.UpdateNEP5Asset(nep5)
			db.InsertNewAsset(nep5)
		}
	}
}
