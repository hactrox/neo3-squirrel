package tasks

import (
	"neo3-squirrel/rpc"
	"neo3-squirrel/tasks/applog"
	"neo3-squirrel/tasks/block"
	"neo3-squirrel/tasks/nep5"
	"neo3-squirrel/util/log"
)

// Run manages all sync tasks.
func Run() {
	log.Info("Start Neo3 blockchain data parser.")

	rpc.TraceBestHeight()

	block.StartBlockSyncTask()
	applog.StartApplicationLogSyncTask()
	nep5.StartNEP5TransferSyncTask()
}
