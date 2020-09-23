package contract

import (
	"neo3-squirrel/cache/asset"
	"neo3-squirrel/cache/contractstate"
	"neo3-squirrel/db"
	"neo3-squirrel/models"
	"neo3-squirrel/rpc"
	"sort"
	"time"
)

const batches = 50

// StartContractTask starts contract related tasks.
func StartContractTask() {
	latestRecord := db.GetLastContractStateRecord()
	fromBlockIndex := uint(0)
	if latestRecord != nil {
		fromBlockIndex = latestRecord.BlockIndex + 1
	}

	go fetchContractStates(fromBlockIndex)
}

func fetchContractStates(fromBlockIndex uint) {
	for {
		contractStates := rpc.GetContractStates(fromBlockIndex, batches)
		l := len(contractStates)
		if l == 0 {
			time.Sleep(3 * time.Second)
			continue
		}

		sort.SliceStable(contractStates, func(i, j int) bool {
			return contractStates[i].BlockIndex < contractStates[j].BlockIndex
		})

		// for _, cs := range contractStates {
		// 	log.Debugf("%s %s %s %d %s", cs.Hash, cs.Name, cs.Symbol, cs.Decimals, convert.BigFloatToString(cs.TotalSupply))
		// }

		list := []*models.ContractState{}

		// Split by block index.
		for i := 0; i < l; i++ {
			cs := contractStates[i]
			list = append(list, models.ParseContractState(cs))
			if i+1 >= l || contractStates[i+1].BlockIndex != cs.BlockIndex {
				contractstate.AddContractState(list)

				// Reset list.
				list = []*models.ContractState{}

				nep5 := &models.Asset{
					BlockIndex:  cs.BlockIndex,
					BlockTime:   cs.BlockTime,
					Contract:    cs.Hash,
					Name:        cs.Name,
					Symbol:      cs.Symbol,
					Decimals:    cs.Decimals,
					Type:        "nep5",
					TotalSupply: cs.TotalSupply,
				}

				asset.UpdateNEP5Asset(nep5)
			}
		}

		fromBlockIndex = contractStates[l-1].BlockIndex + 1
		time.Sleep(100 * time.Millisecond)
	}
}
