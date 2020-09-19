package nep5

import (
	"encoding/base64"
	"fmt"
	"neo3-squirrel/cache/asset"
	"neo3-squirrel/cache/block"
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

		// There may be a time gap between
		// GAS balance updates and transaction gas fee deduction,
		// If transaction sender's GAS changed, task should wait
		// for some time to make sure GAS balance changing was finalized.
		slept := false

		addrTransferCntDelta := countAddressTransfers(txTransfers.transfers)
		committeeGASBalances := getCommitteeGASBalances(txTransfers.transfers)

		for _, transfer := range txTransfers.transfers {
			minBlockIndex := transfer.BlockIndex
			contract := transfer.Contract
			decimals, ok := asset.GetNEP5Decimals(contract)
			if !ok {
				continue
			}
			addrs := map[string]bool{}
			if len(transfer.From) > 0 {
				addrs[transfer.From] = true
			}
			if len(transfer.To) > 0 {
				addrs[transfer.To] = true
			}

			// Get contract balance of these addresses.
			for addr := range addrs {
				// Filter this query if already queried.
				if addrAsset, ok := addrTransfers[addr+contract]; ok {
					addrAsset.Transfers++
					continue
				}

				sleepIfGasConsumed(&slept, minBlockIndex, transfer.TxID, contract, addr)

				balance, ok := util.QueryNEP5Balance(minBlockIndex, addr, contract, decimals)
				if !ok {
					continue
				}

				addrAsset := models.AddrAsset{
					Address:   addr,
					Contract:  contract,
					Balance:   balance,
					Transfers: 1, // Number of transfers added.
				}

				addrAssets = append(addrAssets, &addrAsset)
				addrTransfers[addr+contract] = &addrAsset
			}
		}

		// if new GAS total supply is not nil, then the value should be updated.
		newGASTotalSupply := updateGASTotalSupply(txTransfers.transfers)

		if len(txTransfers.transfers) > 0 ||
			len(addrAssets) > 0 {
			db.InsertNEP5Transfers(txTransfers.transfers,
				addrAssets,
				addrTransferCntDelta,
				newGASTotalSupply,
				committeeGASBalances)
			showTransfers(txTransfers.transfers)
		}
	}
}

func sleepIfGasConsumed(slept *bool, minBlockIndex uint, txID, contract, addr string) {
	if *slept || contract != models.GAS ||
		int(minBlockIndex) != rpc.GetBestHeight() {
		return
	}

	// Check if `from` address is the transaction sender address.
	tx, ok := block.GetTransaction(txID)
	if !ok {
		tx = db.GetTransaction(txID)
	}
	if tx == nil {
		err := fmt.Errorf("failed to find transaction %s", txID)
		log.Panic(err)
	}

	if addr == tx.Sender {
		time.Sleep(1 * time.Second)
		*slept = true
	}
}

func showTransfers(transfers []*models.Transfer) {
	for _, transfer := range transfers {
		from := transfer.From
		to := transfer.To
		amount := transfer.Amount
		contract := transfer.Contract
		symbol, ok := asset.GetNEP5Symbol(contract)
		if !ok {
			log.Panicf("Failed to get asset info of contract %s", contract)
		}

		msg := ""
		amountStr := convert.BigFloatToString(amount)

		if len(from) == 0 {
			// Claim GAS.
			if contract == models.GAS {
				msg = fmt.Sprintf("GAS claimed: %s %s -> %s", amountStr, symbol, to)
			} else {
				msg = fmt.Sprintf("Mint token: %s %s -> %s", amountStr, symbol, to)
			}
		} else {
			if len(to) == 0 {
				msg = fmt.Sprintf("Destroy token: %s destroyed %s %s", from, amount, symbol)
			} else {
				msg = fmt.Sprintf("NEP5 transfer: %34s -> %34s, amount: %s %s",
					from, to, amountStr, symbol)
			}
		}

		log.Info(color.BLightCyanf(msg))
	}
}

func persistExtraAddrBalancesIfExists(noti *models.Notification) bool {
	if util.VMStateFault(noti.VMState) {
		log.Debugf("VM execution status FAULT: %s", noti.TxID)
		return false
	}

	if noti.State == nil ||
		len(noti.State.Value) == 0 {
		return false
	}

	addrAssets := []*models.AddrAsset{}

	for _, stackItem := range noti.State.Value {
		if stackItem.Type != "ByteString" {
			continue
		}

		value := stackItem.Value.(string)
		bytes, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			panic(err)
		}

		if len(bytes) != 20 {
			continue
		}

		addr, ok := util.ExtractAddressFromByteString(value)
		if !ok {
			continue
		}

		contract := noti.Contract

		decimals, ok := asset.GetNEP5Decimals(contract)
		if !ok {
			continue
		}

		balance, ok := util.QueryNEP5Balance(noti.BlockIndex, addr, contract, decimals)
		if !ok {
			continue
		}

		addrAssets = append(addrAssets, &models.AddrAsset{
			Address:   addr,
			Contract:  contract,
			Balance:   balance,
			Transfers: 0,
		})
	}

	if len(addrAssets) > 0 {
		db.PersistNEP5Balances(addrAssets)
	}

	return len(addrAssets) > 0
}
