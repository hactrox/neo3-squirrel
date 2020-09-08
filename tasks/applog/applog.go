package applog

import (
	"fmt"
	"math/big"
	"neo3-squirrel/db"
	"neo3-squirrel/models"
	"neo3-squirrel/rpc"
	"neo3-squirrel/tasks/util"
	"neo3-squirrel/util/color"
	"neo3-squirrel/util/log"
	"sync"
	"time"
)

const chanSize = 5000

var (
	// appLogs stores txID with its applocationlog rpc query result.
	appLogs sync.Map
)

type appLogResult struct {
	tx                *models.Transaction
	appLogQueryResult *rpc.ApplicationLogResult
}

// StartApplicationLogSyncTask starts application log sync task.
func StartApplicationLogSyncTask() {
	lastAppLogTx := db.GetLastTxForApplicationLogTask()

	nextTxPK := uint(0)
	upToBlockHeight := uint(0)
	upToBlockTime := ""

	if lastAppLogTx != nil {
		nextTxPK = lastAppLogTx.ID + 1
		upToBlockHeight = lastAppLogTx.BlockIndex
		if upToBlockHeight > 0 {
			upToBlockTime = time.Unix(int64(lastAppLogTx.BlockTime)/1000, 0).Format("(2006-01-02 15:04:05)")
		}
	}

	remainingTxs := db.GetTxCount(nextTxPK)

	msgs := []string{
		fmt.Sprintf("%s: %s", color.Green("Up to block index"), color.BGreenf("%d%s", upToBlockHeight, upToBlockTime)),
		fmt.Sprintf("%s: %s", color.Green("Transactions left"), color.BGreenf("%d", remainingTxs)),
	}
	log.Info(color.Green("Application log sync progress:"))
	for _, msg := range msgs {
		log.Info("* " + msg)
	}

	// Starts task.

	preAppLogChan := make(chan *models.Transaction, chanSize)
	appLogChan := make(chan *appLogResult, chanSize)

	go fetchApplicationLogs(nextTxPK, preAppLogChan, appLogChan)
	go queryAppLog(3, preAppLogChan)

	go handleApplicationLogs(appLogChan)
}

func fetchApplicationLogs(nextTxPK uint, preAppLogChan chan<- *models.Transaction, appLogChan chan<- *appLogResult) {
	// TODO: mail alert

	for {
		txs := db.GetTransactions(nextTxPK, 1000)
		if len(txs) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		// Send transactions to applog channel
		// and waiting for the applog query result.
		for _, tx := range txs {
			preAppLogChan <- tx
			log.Debugf("send tx %s to preAppLogChan", tx.Hash)
		}

		nextTxPK = txs[len(txs)-1].ID + 1

		for _, tx := range txs {
			for {
				// Get applicationlog from
				result, ok := appLogs.Load(tx.Hash)
				if !ok {
					time.Sleep(10 * time.Millisecond)
					continue
				}

				appLogs.Delete(tx.Hash)

				appLog := appLogResult{
					tx:                tx,
					appLogQueryResult: result.(*rpc.ApplicationLogResult),
				}
				appLogChan <- &appLog
				log.Debugf("send appLog of tx %s into appLogChan", tx.Hash)
				break
			}
		}
	}
}

func queryAppLog(workers int, preAppLogChan <-chan *models.Transaction) {
	// TODO: mail alert

	for i := 0; i < workers; i++ {
		go func(ch <-chan *models.Transaction) {
			for tx := range ch {
				// log.Debugf("prepare to query applicationlog of tx %s", tx.Hash)
				appLogQueryResult := rpc.GetApplicationLog(tx.BlockIndex, tx.Hash)
				appLogs.Store(tx.Hash, appLogQueryResult)
				log.Debugf("store applog result into appLogs, txID=%s, len(noti)=%d", tx.Hash, len(appLogQueryResult.Notifications))
			}
		}(preAppLogChan)
	}
}

func handleApplicationLogs(appLogChan <-chan *appLogResult) {
	// TODO: mail alert

	for result := range appLogChan {
		tx := result.tx
		logResult := result.appLogQueryResult
		log.Debugf("handle application log of txID: %s", tx.Hash)

		// Store applicationlog result
		appLog := models.ParseApplicationLog(tx, logResult)
		appLog.GasConsumed = util.GetReadableAmount(appLog.GasConsumed, big.NewFloat(8))
		db.InsertApplicationLog(&appLog)
	}
}
