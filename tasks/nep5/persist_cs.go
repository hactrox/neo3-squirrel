package nep5

import (
	"neo3-squirrel/cache/contractstate"
	"neo3-squirrel/db"
	"neo3-squirrel/models"
	"neo3-squirrel/util/log"
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
		case models.CSTrackStateAdded:
			added = append(added, cs)
		case models.CSTrackStateUpdated:
			// Passed.
		case models.CSTrackStateDeleted:
			// If contract migration.
			if i-1 >= 0 && csList[i-1].TxID == cs.TxID && csList[i-1].State == models.CSTrackStateAdded {
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
