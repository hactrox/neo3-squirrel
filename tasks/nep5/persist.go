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
	"neo3-squirrel/util/log"
	"time"
)

func persistNEP5Transfers(transferChan <-chan *notiTransfer) {
	lastBlockIndex := uint(0)
	for txTransfers := range transferChan {
		if txTransfers.BlockIndex > lastBlockIndex {
			lastBlockIndex = txTransfers.BlockIndex

			// Handle contract add, migrate, delete actions.
			handleContractStateChange(txTransfers.BlockIndex)
		}

		processNEP5Transfers(txTransfers)
		LastTxPK = txTransfers.PK
	}
}

func processNEP5Transfers(txTransfers *notiTransfer) {
	if len(txTransfers.transfers) == 0 {
		return
	}

	addrAssets := []*models.AddrAsset{}

	// There may be a time gap between
	// GAS balance updates and transaction gas fee deduction,
	// If transaction sender's GAS changed, task should wait
	// for some time to make sure GAS balance changing was finalized.
	slept := false

	addrTransferCntDelta := countAddressTransfers(txTransfers.transfers)

	for _, transfer := range txTransfers.transfers {
		minBlockIndex := transfer.BlockIndex
		contract := transfer.Contract
		decimals, ok := asset.GetNEP5Decimals(contract)
		if !ok {
			continue
		}

		addrs := []string{}
		if len(transfer.From) > 0 {
			addrs = append(addrs, transfer.From)
		}
		if len(transfer.To) > 0 && transfer.To != transfer.From {
			addrs = append(addrs, transfer.To)
		}

		if len(addrs) == 0 {
			return
		}

		readableBalances, ok := util.QueryNEP5Balances(minBlockIndex, addrs, contract, decimals)
		if !ok {
			return
		}

		addrAssetBalanceCache := make(map[string]*models.AddrAsset)

		// Get contract balance of these addresses.
		for idx, addr := range addrs {
			// Filter this query if already queried.
			if addrAsset, ok := addrAssetBalanceCache[addr+contract]; ok {
				addrAsset.Transfers++
				continue
			}

			sleepIfGasConsumed(&slept, minBlockIndex, transfer, contract, addr)

			addrAsset := models.AddrAsset{
				Address:   addr,
				Contract:  contract,
				Balance:   readableBalances[idx],
				Transfers: 1, // Number of transfers added.
			}

			addrAssets = append(addrAssets, &addrAsset)
			addrAssetBalanceCache[addr+contract] = &addrAsset
		}
	}

	// if new GAS total supply is not nil, then the value should be updated.
	newGASTotalSupply := updateGASTotalSupply(txTransfers.transfers)

	if len(txTransfers.transfers) > 0 || len(addrAssets) > 0 {
		db.InsertNEP5Transfers(
			txTransfers.transfers,
			addrAssets,
			addrTransferCntDelta,
			newGASTotalSupply)

		showTransfers(txTransfers.transfers)
	}
}

func sleepIfGasConsumed(slept *bool, minBlockIndex uint, transfer *models.Transfer, contract, addr string) {
	txID := transfer.TxID

	if *slept || contract != models.GAS ||
		int(minBlockIndex) != rpc.GetBestHeight() {
		return
	}

	if transfer.Src == "block" {
		time.Sleep(1 * time.Second)
		*slept = true
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

		readableBalance, ok := util.QueryNEP5Balance(noti.BlockIndex, addr, contract, decimals)
		if !ok {
			continue
		}

		addrAssets = append(addrAssets, &models.AddrAsset{
			Address:   addr,
			Contract:  contract,
			Balance:   readableBalance,
			Transfers: 0,
		})
	}

	if len(addrAssets) > 0 {
		db.PersistNEP5Balances(addrAssets)
	}

	return len(addrAssets) > 0
}
