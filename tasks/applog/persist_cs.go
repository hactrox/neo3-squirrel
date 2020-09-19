package applog

import (
	"encoding/json"
	"math/big"
	"neo3-squirrel/cache/asset"
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

func handleContractStateChange(csList []*models.ContractState) {
	contractStates := []*models.ContractState{}
	added := []*models.ContractState{}
	deleted := []*models.ContractState{}
	migrated := map[*models.ContractState]*models.ContractState{} // map[new]old

	for i := 0; i < len(csList); i++ {
		cs := csList[i]
		// log.Debugf(cs.Hash)
		if cs.TotalSupply == nil {
			cs.TotalSupply = big.NewFloat(0)
		}

		contractStates = append(contractStates, cs)

		switch cs.State {
		case Added:
			updateIfNEP5(cs)

			// If migration.
			if i+1 < len(csList) &&
				csList[i+1].TxID == cs.TxID &&
				csList[i+1].State == Deleted {
				migrated[cs] = csList[i+1]
				i++
				continue
			}

			added = append(added, cs)
		case Updated:
			// Passed.
		case Deleted:
			deleted = append(deleted, cs)
		default:
			log.Panicf("Unsupported contract track state: %s", cs.State)
		}
	}

	db.InsertNewContractStates(contractStates, added, deleted, migrated)
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
			asset.UpdateNEP5Asset(&models.Asset{
				BlockIndex:  c.BlockIndex,
				BlockTime:   c.BlockTime,
				Contract:    c.Hash,
				Name:        c.Name,
				Symbol:      c.Symbol,
				Decimals:    c.Decimals,
				Type:        "nep5",
				TotalSupply: c.TotalSupply,
			})
		}
	}
}
