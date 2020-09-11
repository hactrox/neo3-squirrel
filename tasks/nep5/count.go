package nep5

import (
	"neo3-squirrel/cache/address"
	"neo3-squirrel/models"
)

func countAddressTransfers(transfers []*models.Transfer) map[string]*models.AddressInfo {
	addrInfoUpdates := map[string]*models.AddressInfo{}

	for _, transfer := range transfers {
		from := transfer.From
		to := transfer.To
		blockTime := transfer.BlockTime

		addresses := map[string]bool{}
		if from != "" {
			addresses[from] = true
		}
		if to != "" {
			addresses[to] = true
		}

		for addr := range addresses {
			created := address.Cache(addr, blockTime)
			if created {
				addrInfoUpdates[addr] = &models.AddressInfo{
					Address:     addr,
					FirstTxTime: blockTime,
					LastTxTime:  blockTime,
					Transfers:   1,
				}

				continue
			}

			addrInfo, ok := addrInfoUpdates[addr]
			if ok {
				addrInfo.Transfers++
				continue
			}

			addrInfoUpdates[addr] = &models.AddressInfo{
				Address:    addr,
				LastTxTime: blockTime,
				Transfers:  1,
			}
		}
	}

	return addrInfoUpdates
}
