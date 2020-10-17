package contract

import (
	"neo3-squirrel/cache/asset"
	"neo3-squirrel/cache/contractstate"
	"neo3-squirrel/db"
	"neo3-squirrel/models"
	"neo3-squirrel/rpc"
	"sort"
	"strings"
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
			time.Sleep(1 * time.Second)

			rpcBestIndex := rpc.GetBestHeight()
			for rpcBestIndex >= 0 && rpcBestIndex < int(fromBlockIndex) {
				time.Sleep(100 * time.Millisecond)
				rpcBestIndex = rpc.GetBestHeight()
			}

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
			cs := models.ParseContractState(contractStates[i])

			if cs.State == models.CSTrackStateAdded {
				updateIfNEP5(cs)
			}

			list = append(list, cs)
			if i+1 >= l || contractStates[i+1].BlockIndex != cs.BlockIndex {
				contractstate.AddContractState(list)

				// Reset list.
				list = []*models.ContractState{}
			}
		}

		fromBlockIndex = contractStates[l-1].BlockIndex + 1
		time.Sleep(100 * time.Millisecond)
	}
}

func updateIfNEP5(c *models.ContractState) {
	if c.Name == "" || c.Symbol == "" || len(c.Manifest) == 0 {
		return
	}

	manifestStr := strings.ToLower(string(c.Manifest))

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
