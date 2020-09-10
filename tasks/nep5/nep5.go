package nep5

import (
	"encoding/base64"
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
	"strings"
	"time"
)

var (
	chanSize = 5000

	// nep5Assets loads all known NEP5 assets from DB.
	nep5Assets = map[string]*models.Asset{}
)

type notiTransfer struct {
	TxID      string
	transfers []*models.Transfer
}

// StartNEP5TransferSyncTask starts NEP5 transfer related tasks.
func StartNEP5TransferSyncTask() {
	// Load all known assets from DB.
	assets := db.GetAllAssets("nep5")
	for _, asset := range assets {
		nep5Assets[asset.Contract] = asset
	}

	lastTransferNoti := db.GetLastNotiForNEP5Task()
	upToBlockHeight := uint(0)
	upToBlockTime := ""
	remainingNotis := uint(0)
	lastNotiTxID := ""

	if lastTransferNoti != nil {
		tx := db.GetTransaction(lastTransferNoti.TxID)
		if tx == nil {
			log.Panicf("Failed to get tx detail of tx %s", lastTransferNoti.TxID)
		}

		upToBlockHeight = tx.BlockIndex
		if upToBlockHeight > 0 {
			upToBlockTime = time.Unix(int64(tx.BlockTime/1000), 0).Format("(2006-01-02 15:04:05)")
		}

		remainingNotis = db.GetNotificationCount(lastTransferNoti.ID + 1)
		lastNotiTxID = lastTransferNoti.TxID
	} else {
		remainingNotis = db.GetNotificationCount(0)
	}

	msgs := []string{
		fmt.Sprintf("%s: %s", color.Green("Up to block index"), color.BGreenf("%d%s", upToBlockHeight, upToBlockTime)),
		fmt.Sprintf("%s: %s", color.Green("Notification left"), color.BGreenf("%d", remainingNotis)),
	}
	log.Info(color.Green("NEP5 transfer sync progress:"))
	for _, msg := range msgs {
		log.Info("* " + msg)
	}

	// Starts task.

	transferChan := make(chan *notiTransfer, chanSize)

	go fetchNotifications(lastNotiTxID, transferChan)
	go persistNEP5Transfers(transferChan)
}

func fetchNotifications(lastNotiTxID string, transferChan chan<- *notiTransfer) {
	// TODO: mail alert

	lastApplogID := uint(0)

	if lastNotiTxID != "" {
		appLog := db.GetApplicationLogByTxID(lastNotiTxID)
		if appLog == nil {
			log.Panicf("Failed to get application log of tx %s", lastNotiTxID)
		}

		lastApplogID = appLog.ID
	}

	for {
		notis := db.GetGroupedAppLogNotifications(lastApplogID+1, 100)
		if len(notis) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		// Group notifications by txID.
		notiArrays := [][]*models.Notification{}
		arrIndex := 0
		indexTxID := notis[0].TxID
		notiArrays = append(notiArrays, []*models.Notification{notis[0]})
		for i := 1; i < len(notis); i++ {
			noti := notis[i]
			if noti.TxID != indexTxID {
				arrIndex++
				notiArrays = append(notiArrays, []*models.Notification{noti})
				indexTxID = noti.TxID
				continue
			}

			notiArrays[arrIndex] = append(notiArrays[arrIndex], noti)
		}

		for _, notis := range notiArrays {
			txID := notis[0].TxID
			txTransfers := notiTransfer{TxID: txID}

			for _, noti := range notis {
				switch strings.ToLower(noti.EventName) {
				case "transfer":
					log.Debugf("New NEP5 transfer event detected: %s", noti.TxID)
					transfer := parseNEP5Transfer(noti)
					if transfer != nil {
						txTransfers.transfers = append(txTransfers.transfers, transfer)
					}
				default:
					// Detect if has address parameter, if true, check if has balance.
					if !persistExtraAddrBalancesIfExists(noti) {
						log.Info("Notification in tx %s not parsed. EventName=%s", txID, noti.EventName)
					}
				}
			}

			transferChan <- &txTransfers
		}

		lastTxID := notis[len(notis)-1].TxID
		appLog := db.GetApplicationLogByTxID(lastTxID)
		if appLog == nil {
			log.Panicf("Failed to get application log of tx %s", lastTxID)
		}

		lastApplogID = appLog.ID
	}
}

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

func parseNEP5Transfer(noti *models.Notification) *models.Transfer {
	if util.VMStateFault(noti.VMState) {
		log.Debugf("VM execution status FAULT: %s", noti.TxID)
		return nil
	}

	if noti.State == nil ||
		noti.State.Type != "Array" ||
		len(noti.State.Value) != 3 {
		log.Debug("NEP5 transfer notification state not correct")
		return nil
	}

	// Get contract info.
	contract := noti.Contract
	var asset *models.Asset
	if nep5Asset, ok := nep5Assets[contract]; ok {
		asset = nep5Asset
	} else {
		asset = queryNEP5AssetInfo(noti, contract)
		nep5Assets[asset.Contract] = asset
	}

	if asset == nil {
		log.Debugf("Cannot find asset info for contract: %s", contract)
		return nil
	}

	// Parse Transfer Info.
	stackItems := noti.State.Value
	from, to, amount, ok := util.ExtractNEP5Transfer(stackItems)
	if !ok {
		log.Debug("Failed to extract NEP5 transfer parameters")
		return nil
	}

	readableAmount := convert.AmountReadable(amount, asset.Decimals)

	transfer := models.Transfer{
		BlockIndex: noti.BlockIndex,
		BlockTime:  noti.BlockTime,
		TxID:       noti.TxID,
		Contract:   noti.Contract,
		From:       from,
		To:         to,
		Amount:     readableAmount,
	}

	log.Debug("New NEP5 transfer parsed")
	return &transfer
}

func queryNEP5AssetInfo(noti *models.Notification, contract string) *models.Asset {
	minBlockIndex := noti.BlockIndex
	asset := &models.Asset{
		BlockIndex: minBlockIndex,
		BlockTime:  noti.BlockTime,
		Contract:   contract,
		Type:       "nep5",
	}

	bestBlockIndex := rpc.GetBestHeight()
	util.QueryAssetBasicInfo(minBlockIndex, asset)

	// log.Debugf("Name=%s, Symbol=%s, Decimals=%v, TotalSupply=%v", name, symbol, decimals, totalSupply)
	db.InsertNewAsset(asset)

	if asset.Contract == models.GAS && bestBlockIndex > 0 {
		gas.CacheGASTotalSupply(uint(bestBlockIndex), asset.TotalSupply)
	}

	return asset
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

		contractBalance, ok := util.QueryNEP5Balance(noti.BlockIndex, addr, contract)
		if !ok {
			continue
		}

		addrAssets = append(addrAssets, &models.AddrAsset{
			Address:   addr,
			Contract:  contract,
			Balance:   contractBalance,
			Transfers: 0,
		})
	}

	if len(addrAssets) > 0 {
		db.PersistNEP5Balances(addrAssets)
	}

	return len(addrAssets) > 0
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
