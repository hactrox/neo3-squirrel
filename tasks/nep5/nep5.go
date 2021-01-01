package nep5

import (
	"fmt"
	"neo3-squirrel/cache/asset"
	"neo3-squirrel/cache/gas"
	"neo3-squirrel/config"
	"neo3-squirrel/db"
	"neo3-squirrel/models"
	"neo3-squirrel/rpc"
	"neo3-squirrel/tasks/util"
	"neo3-squirrel/util/color"
	"neo3-squirrel/util/convert"
	"neo3-squirrel/util/log"
	"neo3-squirrel/util/timeutil"
	"strings"
	"time"
)

var (
	chanSize = 8000

	// LastTxBlockIndex is the block index of the last transfer.
	LastTxBlockIndex uint
)

type notiTransfer struct {
	// Hash may be block hash or transaction hash.
	Hash       string
	BlockIndex uint
	transfers  []*models.Transfer
}

// StartNEP5TransferSyncTask starts NEP5 transfer related tasks.
func StartNEP5TransferSyncTask() {
	lastTransferNoti := db.GetLastNotiForNEP5Task()
	upToBlockHeight := uint(0)
	upToBlockTime := ""
	remainingNotis := uint(0)
	lastNotiPK := uint(0)

	if lastTransferNoti != nil {
		upToBlockHeight = lastTransferNoti.BlockIndex
		if upToBlockHeight > 0 {
			upToBlockTime = fmt.Sprintf("(%s)", timeutil.FormatBlockTime(lastTransferNoti.BlockTime))
		}

		lastNotiPK = lastTransferNoti.ID
		remainingNotis = db.GetNotificationCount(lastTransferNoti.ID + 1)
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

	// Starts tasks.
	transferChan := make(chan *notiTransfer, chanSize)

	go fetchNotifications(lastNotiPK+1, transferChan)
	go persistNEP5Transfers(transferChan)
}

func fetchNotifications(nextNotiPK uint, transferChan chan<- *notiTransfer) {
	for {
		notis := db.GetNotificationsGroupedByHash(nextNotiPK, 100)
		if len(notis) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		// Group notifications by hash.
		notiArrays := groupNotiByHash(notis)

		// Every notiArrat has the same hash.
		for _, notis := range notiArrays {
			// hash and blockIndex are the same across these grouped notis,
			// so get them from the first noti.
			hash := notis[0].Hash
			blockIndex := notis[0].BlockIndex

			transferInfo := notiTransfer{
				Hash:       hash,
				BlockIndex: blockIndex,
			}

			for _, noti := range notis {
				eventName := noti.EventName

				switch strings.ToLower(eventName) {
				case "transfer":
					log.Debugf("New NEP5 transfer event detected: %s", hash)
					transfer := parseNEP5Transfer(noti)
					if transfer != nil {
						transferInfo.transfers = append(transferInfo.transfers, transfer)
					}
				case strings.ToLower(string(models.ContractDestroyEvent)):
					handleAssetDestroy(noti)
				default:
					// Detect if has address parameter, if true, check if has balance.
					if !persistExtraAddrBalancesIfExists(noti) {
						log.Infof("Notification in hash %s(%s) not parsed. EventName=%s", hash, noti.Src, eventName)
					}
				}
			}

			transferChan <- &transferInfo
		}

		nextNotiPK = notis[len(notis)-1].ID + 1
	}
}

func groupNotiByHash(notis []*models.Notification) [][]*models.Notification {
	notiArrays := [][]*models.Notification{}
	arrIndex := 0
	indexHash := notis[0].Hash
	notiArrays = append(notiArrays, []*models.Notification{notis[0]})
	for i := 1; i < len(notis); i++ {
		noti := notis[i]
		if noti.Hash != indexHash {
			arrIndex++
			notiArrays = append(notiArrays, []*models.Notification{noti})
			indexHash = noti.Hash
			continue
		}

		notiArrays[arrIndex] = append(notiArrays[arrIndex], noti)
	}

	return notiArrays
}

func handleAssetDestroy(noti *models.Notification) {
	contractHash, ok := util.GetContractHash(noti)
	if !ok {
		return
	}

	db.DestroyAsset(contractHash)
	asset.Remove(contractHash)
}

func parseNEP5Transfer(noti *models.Notification) *models.Transfer {
	if util.VMStateFault(noti.VMState) {
		log.Debugf("VM execution status FAULT: %s", noti.Hash)
		return nil
	}

	if noti.State == nil ||
		noti.State.Type != "Array" ||
		len(noti.State.Value) != 3 {
		log.Debug("NEP5 transfer notification state not correct")
		return nil
	}

	// Get contract info.
	assetHash := noti.Contract
	decimals, ok := asset.GetDecimals(assetHash)
	if !ok {
		nep5 := queryNEP5AssetInfo(noti, assetHash)
		if nep5 == nil {
			return nil
		}

		db.InsertNewAsset(nep5)

		asset.Update(nep5)
		decimals = nep5.Decimals
	}

	// Parse Transfer Info.
	stackItems := noti.State.Value
	from, to, rawAmount, ok := util.ExtractNEP5Transfer(stackItems)
	if !ok {
		log.Debug("Failed to extract NEP5 transfer parameters")
		return nil
	}

	readableAmount := convert.AmountReadable(rawAmount, decimals)

	transfer := models.Transfer{
		BlockIndex: noti.BlockIndex,
		BlockTime:  noti.BlockTime,
		Hash:       noti.Hash,
		Src:        noti.Src,
		Contract:   noti.Contract,
		From:       from,
		To:         to,
		Amount:     readableAmount,
	}

	if readableAmount.Cmp(config.MaxVal) > 0 {
		log.Errorf("NEP5 transfer amount exceed, transfer skipped: %v", transfer)
		return nil
	}

	return &transfer
}

func queryNEP5AssetInfo(noti *models.Notification, contract string) *models.Asset {
	minBlockIndex := noti.BlockIndex
	asset := models.Asset{
		BlockIndex: minBlockIndex,
		BlockTime:  noti.BlockTime,
		TxID:       noti.Hash,
		Contract:   contract,
	}

	// Get and save contract manifest.
	contractState := rpc.GetContractState(noti.BlockIndex, contract)
	asset.Name = contractState.Manifest.Name

	bestBlockIndex := rpc.GetBestHeight()
	ok := util.QueryAssetBasicInfo(minBlockIndex, &asset)
	if !ok {
		log.Warnf("Failed to get NEP5 contract info. Hash=%s(%s), Contract=%s, BlockIndex=%d, BlockTime=%s",
			noti.Hash, noti.Src, contract, noti.BlockIndex, timeutil.FormatBlockTime(noti.BlockTime))
		return nil
	}

	if contract == models.GAS && bestBlockIndex > 0 {
		gas.CacheGASTotalSupply(uint(bestBlockIndex), asset.TotalSupply)
	}

	if asset.TotalSupply != nil && asset.TotalSupply.Cmp(config.MaxVal) > 0 {
		log.Errorf("Asset total supply , asset skipped: %v", asset)
	}

	return &asset
}
