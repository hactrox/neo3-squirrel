package applog

import (
	"fmt"
	"neo3-squirrel/cache/block"
	"neo3-squirrel/db"
	"neo3-squirrel/models"
	"neo3-squirrel/rpc"
	"neo3-squirrel/util/color"
	"neo3-squirrel/util/log"
	"neo3-squirrel/util/timeutil"
	"sync"
	"time"
)

const chanSize = 5000

var (
	// appLogs stores txID with its applocationlog rpc query result.
	appLogs sync.Map

	// LastAppLogBlockIndex is the highest applog block index persisted.
	LastAppLogBlockIndex uint

	preAppLogChan = make(chan *preAppLog, chanSize)
	appLogChan    = make(chan *appLogInfo, chanSize)

	queryResults = []*appLogInfo{}
)

type preAppLog struct {
	BlockIndex uint
	Hash       string
}

type appLogInfo struct {
	BlockIndex uint
	BlockTime  uint64
	Hash       string
	appLog     *rpc.ApplicationLog
}

// StartApplicationLogSyncTask starts application log sync task.
func StartApplicationLogSyncTask() {
	lastNoti := db.GetLastNotification()

	if lastNoti != nil {
		upToBlockTime := fmt.Sprintf("(%s)", timeutil.FormatBlockTime(lastNoti.BlockTime))

		msgs := []string{
			fmt.Sprintf("%s: %s", color.Green("Up to block index"),
				color.BGreenf("%d%s", lastNoti.BlockIndex, upToBlockTime)),
		}

		log.Info(color.Green("Application log sync progress:"))

		for _, msg := range msgs {
			log.Info("* " + msg)
		}
	}

	// Start tasks.
	go fetchApplicationLogs(lastNoti)
	go queryAppLog(3, preAppLogChan)
	go persistApplicationLogs(appLogChan)
}

func fetchApplicationLogs(lastNoti *models.Notification) {
	processLastBlockNotifications(lastNoti)

	nextBlockIndex := uint(0)
	if lastNoti != nil {
		nextBlockIndex = lastNoti.BlockIndex + 1
	}

	for {
		block, ok := block.GetBlock(nextBlockIndex)
		if !ok {
			block = db.GetBlock(nextBlockIndex)
			if block == nil {
				time.Sleep(1 * time.Second)
				continue
			}

			block.SetTxs(db.GetBlockTxs(nextBlockIndex))
		}

		nextBlockIndex++

		preAppLogPushBlock(block)

		// Send transactions to pre-applog channel
		// and waiting for the applog query result.
		for _, tx := range block.GetTxs() {
			preAppLogPushTx(tx)
		}

		for i := 0; i < len(queryResults); i++ {
			retry := 0

			for {
				logInfo := queryResults[i]

				// Wait for at most 30 seconds to crash unless all fullnodes down.
				if retry > 30*1000 {
					if rpc.AllFullnodesDown() {
						retry = 0
						for rpc.AllFullnodesDown() {
							time.Sleep(100 * time.Millisecond)
						}
						continue
					}

					log.Panicf("Failed to get applog of %s(index=%d, time=%d)",
						logInfo.Hash, logInfo.BlockIndex, logInfo.BlockTime)
				}

				hash := logInfo.Hash

				// Get applicationlog from query result chan.
				result, ok := appLogs.Load(hash)
				if !ok {
					retry += 5
					time.Sleep(5 * time.Millisecond)
					continue
				}

				appLogs.Delete(hash)

				logInfo.appLog = result.(*rpc.ApplicationLog)

				appLogChan <- logInfo
				break
			}
		}

		// Clear query result array.
		queryResults = []*appLogInfo{}
	}
}

func processLastBlockNotifications(lastNoti *models.Notification) {
	if lastNoti == nil {
		return
	}

	blockIndex := lastNoti.BlockIndex

	block, ok := block.GetBlock(blockIndex)
	if !ok {
		block = db.GetBlock(blockIndex)
		if block == nil {
			log.Panicf("Failed to get block at index %d", blockIndex)
		}

		block.SetTxs(db.GetBlockTxs(blockIndex))
	}

	txs := block.GetTxs()
	if len(txs) == 0 {
		return
	}

	txsToAdd := []*models.Transaction{}

	for i := len(txs) - 1; i >= 0; i-- {
		tx := txs[i]
		if tx.Hash == lastNoti.Hash {
			break
		}

		txsToAdd = append(txsToAdd, tx)
	}

	if len(txsToAdd) == 0 {
		return
	}

	for i := len(txsToAdd) - 1; i >= 0; i-- {
		preAppLogPushTx(txsToAdd[i])
	}
}

func preAppLogPushBlock(block *models.Block) {
	pushToPreChan(block.Index, block.Time, block.Hash)
}

func preAppLogPushTx(tx *models.Transaction) {
	pushToPreChan(tx.BlockIndex, tx.BlockTime, tx.Hash)
}

func pushToPreChan(blockIndex uint, blockTime uint64, hash string) {
	preAppLogChan <- &preAppLog{
		BlockIndex: blockIndex,
		Hash:       hash,
	}

	queryResults = append(queryResults, &appLogInfo{
		BlockIndex: blockIndex,
		BlockTime:  blockTime,
		Hash:       hash,
		appLog:     nil,
	})
}
