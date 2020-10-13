package tasks

import (
	"neo3-squirrel/cache/address"
	"neo3-squirrel/cache/asset"
	"neo3-squirrel/db"
	"neo3-squirrel/rpc"
	"neo3-squirrel/tasks/applog"
	"neo3-squirrel/tasks/block"
	"neo3-squirrel/tasks/contract"
	"neo3-squirrel/tasks/nep5"
	"neo3-squirrel/util/log"
)

// Run manages all sync tasks.
func Run() {
	log.Info("Start Neo3 blockchain data parser.")

	address.Init(db.GetAllAddressInfo())

	// Load all known assets from DB.
	assets := db.GetAllAssets("nep5")
	asset.UpdateNEP5Assets(assets)

	rpc.TraceBestHeight()

	block.StartBlockSyncTask()
	contract.StartContractTask()
	applog.StartApplicationLogSyncTask()
	nep5.StartNEP5TransferSyncTask()
}
