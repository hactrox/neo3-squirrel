package nep5

import (
	"fmt"
	"math/big"
	"neo3-squirrel/cache/block"
	"neo3-squirrel/cache/gas"
	"neo3-squirrel/db"
	"neo3-squirrel/models"
	"neo3-squirrel/rpc"
	"neo3-squirrel/tasks/util"
	"neo3-squirrel/util/color"
	"neo3-squirrel/util/convert"
	"neo3-squirrel/util/log"
	"time"
)

func persistNEP5Transfers(transferChan <-chan *notiTransfer) {
	for txTransfers := range transferChan {
		if len(txTransfers.transfers) == 0 {
			continue
		}

		addrAssets := []*models.AddrAsset{}
		addrTransfers := make(map[string]*models.AddrAsset)

		for _, transfer := range txTransfers.transfers {
			minBlockIndex := transfer.BlockIndex
			contract := transfer.Contract
			asset := nep5Assets[contract]
			addrs := []string{}
			if len(transfer.From) > 0 {
				addrs = append(addrs, transfer.From)
			}
			if len(transfer.To) > 0 {
				addrs = append(addrs, transfer.To)
			}

			for _, addr := range addrs {
				// Filter this query if alreadt queried.
				if addrAsset, ok := addrTransfers[addr+contract]; ok {
					addrAsset.Transfers++
					continue
				}

				if minBlockIndex > 0 &&
					int(minBlockIndex) == rpc.GetBestHeight() &&
					contract == models.GAS {
					// Check if `from` address is the transaction sender address.
					tx, ok := block.GetTransaction(transfer.TxID)
					if !ok {
						tx = db.GetTransaction(transfer.TxID)
					}
					if tx == nil {
						err := fmt.Errorf("failed to find transaction %s", transfer.TxID)
						log.Panic(err)
					}

					if addr == tx.Sender {
						time.Sleep(1 * time.Second)
					}
				}

				amount, ok := util.QueryNEP5Balance(minBlockIndex, addr, contract)
				if !ok {
					continue
				}

				addrAsset := models.AddrAsset{
					Address:   addr,
					Contract:  contract,
					Balance:   convert.AmountReadable(amount, asset.Decimals),
					Transfers: 1, // Number of contract transfers added.
				}

				addrAssets = append(addrAssets, &addrAsset)
				addrTransfers[addr+contract] = &addrAsset
			}
		}

		// if new GAS total supply is not nil, then the value should be updated.
		newGASTotalSupply := updateGASTotalSupply(txTransfers.transfers)

		if len(txTransfers.transfers) > 0 ||
			len(addrAssets) > 0 {
			db.InsertNEP5Transfers(txTransfers.transfers, addrAssets, newGASTotalSupply)
			showTransfers(txTransfers.transfers)
		}
	}
}

func updateGASTotalSupply(transfers []*models.Transfer) *big.Float {
	// Check if has GAS claim transfer.
	hasGASClaimTransfer := false

	for _, transfer := range transfers {
		if transfer.IsGASClaimTransfer() {
			hasGASClaimTransfer = true
			break
		}
	}

	bestBlock := rpc.GetBestHeight()
	if !hasGASClaimTransfer || int(gas.CachedTillBlockIndex()) >= bestBlock || bestBlock < 0 {
		return nil
	}

	gasTotalSupply, ok := util.QueryAssetTotalSupply(uint(bestBlock), models.GAS, 8)
	if !ok {
		return nil
	}

	gas.CacheGASTotalSupply(uint(bestBlock), gasTotalSupply)
	return gasTotalSupply
}

func showTransfers(transfers []*models.Transfer) {
	for _, transfer := range transfers {
		from := transfer.From
		to := transfer.To
		amount := transfer.Amount
		contract := transfer.Contract
		asset, ok := nep5Assets[contract]
		if !ok {
			log.Panicf("Failed to get asset info of contract %s", contract)
		}

		msg := ""
		amountStr := convert.BigFloatToString(amount)

		if len(from) == 0 {
			// Claim GAS.
			if contract == models.GAS {
				msg = fmt.Sprintf("GAS claimed: %s %s -> %s", amountStr, asset.Symbol, to)
			} else {
				msg = fmt.Sprintf("Mint token: %s %s -> %s", amountStr, asset.Symbol, to)
			}
		} else {
			if len(to) == 0 {
				msg = fmt.Sprintf("Destroy token: %s destroyed %s %s", from, amount, asset.Symbol)
			} else {
				msg = fmt.Sprintf("NEP5 transfer: %34s -> %34s, Amount=%s %s",
					from, to, amountStr, asset.Symbol)
			}
		}

		log.Info(color.BLightCyanf(msg))
	}
}
