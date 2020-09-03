package tasks

import (
	"neo3-squirrel/db"
	"neo3-squirrel/rpc"
	"neo3-squirrel/tasks/block"
	"neo3-squirrel/util/color"
	"neo3-squirrel/util/log"
)

// Run manages all sync tasks.
func Run() {
	log.Info("Start Neo3 blockchain data parser.")

	initTask()

	block.StartBlockSyncTask()
}

func initTask() {
	rpc.TraceBestHeight()

	lastBlockHeight := db.GetLastBlockHeight()
	bestBlockHeight := rpc.GetBestHeight()

	log.Info(color.BGreenf("Block sync progress: %d/%d, %d blocks behind",
		lastBlockHeight, bestBlockHeight, bestBlockHeight-lastBlockHeight))
}
