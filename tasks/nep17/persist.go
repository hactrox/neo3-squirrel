package nep17

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

func persistNEP17Transfers(transferChan <-chan *notiTransfer) {
	for txTransfers := range transferChan {
		processNEP17Transfers(txTransfers)
		LastTxBlockIndex = txTransfers.BlockIndex
	}
}

func processNEP17Transfers(txTransfers *notiTransfer) {
	if len(txTransfers.transfers) == 0 {
		return
	}

	addrAssets := []*models.AddrAsset{}
	addrAssetBalanceCache := make(map[string]*models.AddrAsset)

	// There may be a time gap between
	// GAS balance updates and transaction gas fee deduction,
	// If transaction sender's GAS changed, task should wait
	// for some time to make sure GAS balance changing was finalized.
	slept := false

	txAddrInfo := getTxAddrInfo(txTransfers.transfers)

	for _, transfer := range txTransfers.transfers {
		assetHash := transfer.Contract
		decimals, ok := asset.GetDecimals(assetHash)
		if !ok {
			continue
		}

		addrs := getTransferAddrs(transfer)
		if len(addrs) == 0 {
			continue
		}

		minBlockIndex := transfer.BlockIndex
		readableBalances, ok := util.QueryNEP17Balances(minBlockIndex, addrs, assetHash, decimals)
		if !ok {
			continue
		}

		// Get asset balance of these addresses.
		for idx, addr := range addrs {
			// Filter this query if already queried.
			if addrAsset, ok := addrAssetBalanceCache[addr+assetHash]; ok {
				addrAsset.Transfers++
				continue
			}

			sleepIfGasConsumed(&slept, minBlockIndex, transfer, assetHash, addr)

			addrAsset := models.AddrAsset{
				Address:   addr,
				Contract:  assetHash,
				Balance:   readableBalances[idx],
				Transfers: 1, // Number of transfers added.
			}

			addrAssets = append(addrAssets, &addrAsset)
			addrAssetBalanceCache[addr+assetHash] = &addrAsset
		}
	}

	// if new GAS total supply is not nil, then the value should be updated.
	newGASTotalSupply := updateGASTotalSupply(txTransfers.transfers)

	if len(txTransfers.transfers) > 0 || len(addrAssets) > 0 {
		db.InsertNEP17Transfers(
			txTransfers.transfers,
			addrAssets,
			txAddrInfo,
			newGASTotalSupply)

		showTransfers(txTransfers.transfers)
	}
}

func getTransferAddrs(transfer *models.Transfer) []string {
	addrs := []string{}
	if len(transfer.From) > 0 {
		addrs = append(addrs, transfer.From)
	}
	if len(transfer.To) > 0 && transfer.To != transfer.From {
		addrs = append(addrs, transfer.To)
	}

	return addrs
}

func sleepIfGasConsumed(slept *bool, minBlockIndex uint, transfer *models.Transfer, contract, addr string) {
	txID := transfer.Hash

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
		log.Debugf("VM execution status FAULT: %s", noti.Hash)
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

		decimals, ok := asset.GetDecimals(contract)
		if !ok {
			continue
		}

		readableBalance, ok := util.QueryNEP17Balance(noti.BlockIndex, addr, contract, decimals)
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
		db.PersistNEP17Balances(addrAssets)
	}

	return len(addrAssets) > 0
}
