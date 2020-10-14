package applog

import (
	"fmt"
	"neo3-squirrel/cache/block"
	"neo3-squirrel/db"
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
	// LastTxPK if the last tx pk of persisted tx appLogs.
	LastTxPK uint
)

type preAppLog struct {
	BlockIndex uint
	Hash       string
}

type appLogResult struct {
	PK                uint
	BlockIndex        uint
	BlockTime         uint64
	Hash              string
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
			upToBlockTime = fmt.Sprintf("(%s)", timeutil.FormatBlockTime(lastAppLogTx.BlockTime))
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

	preAppLogChan := make(chan *preAppLog, chanSize)
	appLogChan := make(chan *appLogResult, chanSize)

	go fetchApplicationLogs(nextTxPK, preAppLogChan, appLogChan)
	go queryAppLog(3, preAppLogChan)

	go persistApplicationLogs(appLogChan)
}

func fetchApplicationLogs(nextTxPK uint, preAppLogChan chan<- *preAppLog, appLogChan chan<- *appLogResult) {
	// Skip the genesis block.
	nextBlockIndex := uint(1)
	lastBlockAppLog := db.GetLastSystemAppLog()
	if lastBlockAppLog != nil {
		nextBlockIndex = lastBlockAppLog.BlockIndex + 1
	}

	for {
		txs := db.GetTransactions(nextTxPK, 20)
		if len(txs) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		// Prepare app log result.
		queryResult := []*appLogResult{}

		// Send transactions to applog channel
		// and waiting for the applog query result.
		for _, tx := range txs {
			if tx.BlockIndex >= nextBlockIndex {
				nextBlockIndex = tx.BlockIndex + 1
				block, ok := block.GetBlock(tx.BlockIndex)
				if !ok {
					block = db.GetBlock(tx.BlockIndex)
					if block == nil {
						log.Panicf("Failed to get block at index %d", tx.BlockIndex)
					}
				}

				preAppLogChan <- &preAppLog{
					BlockIndex: block.Index,
					Hash:       block.Hash,
				}
				queryResult = append(queryResult, &appLogResult{
					BlockIndex:        block.Index,
					BlockTime:         block.Time,
					Hash:              block.Hash,
					appLogQueryResult: nil,
				})
			}

			preAppLogChan <- &preAppLog{
				BlockIndex: tx.BlockIndex,
				Hash:       tx.Hash,
			}
			queryResult = append(queryResult, &appLogResult{
				PK:                tx.ID,
				BlockIndex:        tx.BlockIndex,
				BlockTime:         tx.BlockTime,
				Hash:              tx.Hash,
				appLogQueryResult: nil,
			})

			// log.Debugf("send tx %s to preAppLogChan", tx.Hash)
		}

		nextTxPK = txs[len(txs)-1].ID + 1

		for i := 0; i < len(queryResult); i++ {
			retry := 0

			for {
				re := queryResult[i]

				// Wait for at most 30 seconds to crash unless all fullnodes down.
				if retry > 3000 {
					if rpc.AllFullnodesDown() {
						retry = 0
						for rpc.AllFullnodesDown() {
							time.Sleep(100 * time.Millisecond)
						}
						continue
					}

					log.Panicf("Failed to get applog of %s(index=%d, time=%d)", re.Hash, re.BlockIndex, re.BlockTime)
				}

				// Get applicationlog from
				result, ok := appLogs.Load(re.Hash)
				if !ok {
					retry++
					time.Sleep(10 * time.Millisecond)
					continue
				}

				appLogs.Delete(re.Hash)

				re.appLogQueryResult = result.(*rpc.ApplicationLogResult)
				re.appLogQueryResult.TxID = re.Hash

				appLogChan <- re
				// log.Debugf("send appLog of tx %s into appLogChan", re.Hash)
				break
			}
		}
	}
}
