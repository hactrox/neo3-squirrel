package tasks

import (
	"neo3-squirrel/db"
	"neo3-squirrel/log"
	"neo3-squirrel/rpc"
	"neo3-squirrel/tasks/block"
	"neo3-squirrel/util/color"
)

func Run() {
	log.Info(color.Green("Start Neo3 blockchain data parser."))

	initTask()

	block.StartBlockSyncTask()
}

func initTask() {
	rpc.TraceBestHeight()

	lastBlockHeight := db.GetLastBlockHeight()
	bestBlockHeight := rpc.GetBestHeight()

	log.Info(color.Greenf("Block sync progress: %d/%d, %d blocks behind",
		lastBlockHeight, bestBlockHeight, bestBlockHeight-lastBlockHeight))
}
