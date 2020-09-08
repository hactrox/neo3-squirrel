package nep5

import (
	"encoding/base64"
	"fmt"
	"math"
	"math/big"
	"neo3-squirrel/db"
	"neo3-squirrel/models"
	"neo3-squirrel/rpc"
	"neo3-squirrel/util/base58"
	"neo3-squirrel/util/color"
	"neo3-squirrel/util/convert"
	"neo3-squirrel/util/log"
	"strconv"
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
					transfer := handleNEP5Transfer(noti)
					if transfer != nil {
						txTransfers.transfers = append(txTransfers.transfers, transfer)
					}
				default:
					log.Info("Notification in tx %s not parsed. EventName=%s", txID, noti.EventName)
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
		// TODO: persist info.
		db.InsertNEP5Transfers(txTransfers.transfers)
	}
}

func handleNEP5Transfer(noti *models.Notification) *models.Transfer {
	if strings.Contains(noti.VMState, "FAULT") {
		log.Debugf("VM execution status FAULT")
		return nil
	}

	if noti.State == nil ||
		noti.State.Type != "Array" ||
		len(noti.State.Value) != 3 {
		return nil
	}

	// Get contract info.
	contract := noti.Contract
	var asset *models.Asset
	if nep5Asset, ok := nep5Assets[contract]; ok {
		asset = nep5Asset
	} else {
		asset = queryNEP5Info(noti, contract)
	}

	if asset == nil {
		log.Debugf("Cannot find asset info for contract: %s", contract)
		return nil
	}

	// Parse Transfer Info.
	stackItems := noti.State.Value
	from, ok := extractAddress(stackItems[0].Type, stackItems[0].Value)
	if !ok {
		return nil
	}
	to, ok := extractAddress(stackItems[1].Type, stackItems[1].Value)
	if !ok {
		return nil
	}
	amount, ok := extractValue(stackItems[2].Type, stackItems[2].Value)
	if !ok {
		return nil
	}

	decimals, _ := asset.Decimals.Int64()
	decimalsFactor := math.Pow10(int(decimals))
	readableAmount := amount.Quo(amount, big.NewFloat(decimalsFactor))
	transferMsg := color.BLightGreenf("NEP5 transfer: %34s -> %34s, Amount=%s %s",
		from, to, convert.BigFloatToString(readableAmount), asset.Symbol)
	log.Info(transferMsg)

	transfer := models.Transfer{
		BlockIndex: noti.BlockIndex,
		BlockTime:  noti.BlockTime,
		TxID:       noti.TxID,
		From:       from,
		To:         to,
		Amount:     readableAmount,
	}

	return &transfer
}

func queryNEP5Info(noti *models.Notification, contract string) *models.Asset {
	minBlockIndex := noti.BlockIndex
	name, ok := queryContractName(minBlockIndex, contract)
	if !ok {
		return nil
	}
	symbol, ok := queryContractSymbol(minBlockIndex, contract)
	if !ok {
		return nil
	}

	decimals, ok := queryContractDecimals(minBlockIndex, contract)
	if !ok {
		return nil
	}

	totalSupply, ok := queryContractTotalSupply(minBlockIndex, contract)
	if !ok {
		return nil
	}

	// log.Debugf("Name=%s, Symbol=%s, Decimals=%v, TotalSupply=%v", name, symbol, decimals, totalSupply)
	return insertNewAsset(minBlockIndex, noti.BlockTime, contract, name, symbol, decimals, totalSupply)
}

func insertNewAsset(blockIndex uint, blockTime uint64, contract, name, symbol string, decimals, totalSupply *big.Float) *models.Asset {
	asset := models.Asset{
		BlockIndex:  blockIndex,
		BlockTime:   blockTime,
		Contract:    contract,
		Name:        name,
		Symbol:      symbol,
		Decimals:    decimals,
		Type:        "nep5",
		TotalSupply: totalSupply,
	}

	db.InsertNewAsset(&asset)
	return &asset
}

func queryContractName(minBlockIndex uint, contract string) (string, bool) {
	stack, ok := queryContractProperty(minBlockIndex, contract, "name")
	if !ok {
		return "", false
	}

	return extractString(stack.Type, stack.Value)
}

func queryContractSymbol(minBlockIndex uint, contract string) (string, bool) {
	stack, ok := queryContractProperty(minBlockIndex, contract, "symbol")
	if !ok {
		return "", false
	}

	return extractString(stack.Type, stack.Value)
}

func queryContractDecimals(minBlockIndex uint, contract string) (*big.Float, bool) {
	stack, ok := queryContractProperty(minBlockIndex, contract, "decimals")
	if !ok {
		return nil, false
	}

	return extractValue(stack.Type, stack.Value)
}

func queryContractTotalSupply(minBlockIndex uint, contract string) (*big.Float, bool) {
	stack, ok := queryContractProperty(minBlockIndex, contract, "totalSupply")
	if !ok {
		return nil, false
	}

	return extractValue(stack.Type, stack.Value)
}

// name, symbol...
func queryContractProperty(minBlockIndex uint, contract, property string) (*rpc.StackItem, bool) {
	result := rpc.InvokeFunction(minBlockIndex, contract, property, nil)
	if result == nil ||
		strings.Contains(result.State, "FAULT") ||
		len(result.Stack) == 0 {
		return nil, false
	}

	return &result.Stack[0], true
}

func extractAddress(typ string, value interface{}) (string, bool) {
	switch typ {
	case "Any":
		return "", true
	case "ByteString":
		bytes, err := base64.StdEncoding.DecodeString(value.(string))
		if err != nil {
			panic(err)
		}

		bytes = append([]byte{0x35}, bytes...)
		return base58.CheckEncode(bytes), true
	default:
		err := fmt.Errorf("failed to parse address in type %s and value=%v", typ, value)
		log.Error(err)
		return "", false
	}
}

func extractString(typ string, value interface{}) (string, bool) {
	switch typ {
	case "Any":
		return "", true
	case "ByteString":
		bytes, err := base64.StdEncoding.DecodeString(value.(string))
		if err != nil {
			log.Error("Failed to extract value(%v) of 'ByteString' type", value)
			return "", false
		}

		return string(bytes), true
	default:
		log.Errorf("Unsupported string extract type: %s, value=%v", typ, value)
		return "", false
	}
}

func extractValue(typ string, value interface{}) (*big.Float, bool) {
	switch typ {
	case "Boolean":
		return big.NewFloat(0), true
	case "Integer":
		valStr := value.(string)
		val, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			log.Error("Failed to extract value(%v) of 'Integer' type", value)
			return nil, false
		}
		return new(big.Float).SetInt64(val), true
	default:
		return nil, false
	}
}
