package nep5

import (
	"fmt"
	"neo3-squirrel/db"
	"neo3-squirrel/models"
	"neo3-squirrel/util/color"
	"neo3-squirrel/util/log"
	"strings"
	"time"
)

var chanSize = 5000

type notiTransfer struct {
	txID      string
	transfers []*models.Transfer
}

func StartNEP5TransferSyncTask() {
	lastTransferNoti := db.GetLastNotiForNEP5Task()
	// nextNotiPK := uint(0)
	upToBlockHeight := uint(0)
	upToBlockTime := ""
	remainingNotis := uint(0)
	lastNotiTxID := ""

	if lastTransferNoti != nil {
		tx := db.GetTransaction(lastTransferNoti.TxID)
		if tx == nil {
			log.Panicf("Failed to get tx detail of tx %s", lastTransferNoti.TxID)
		}

		// nextNotiPK = lastTransferNoti.ID + 1
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
}

func fetchNotifications(lastNotiTxID string, notiChan chan<- *notiTransfer) {
	// TODO: mail alert

	lastApplogID := uint(0)

	if lastNotiTxID == "" {
		appLog := db.GetApplicationLogByID(1)
		if appLog != nil {
			lastApplogID = appLog.ID
		}
	} else {
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
		notiMap := make(map[string][]*models.Notification)
		for _, noti := range notis {
			notiMap[noti.TxID] = append(notiMap[noti.TxID], noti)
		}

		for txID, notis := range notiMap {
			info := notiTransfer{txID: txID}

			for _, noti := range notis {
				switch strings.ToLower(noti.EventName) {
				case "transfer":
					transfer := handleNEP5Transfer(noti)
					if transfer != nil {
						info.transfers = append(info.transfers, transfer)
					}
				default:
					log.Info("Notification in tx %s not parsed. EventName=%s", txID, noti.EventName)
				}
			}
		}

		lastTxID := notis[len(notis)-1].TxID
		appLog := db.GetApplicationLogByTxID(lastTxID)
		if appLog == nil {
			log.Panicf("Failed to get application log of tx %s", lastTxID)
		}

		lastApplogID = appLog.ID
	}
}

func handleNEP5Transfer(noti *models.Notification) *models.Transfer {
	// TODO: implement function.
	return nil
}
