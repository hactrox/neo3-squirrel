package tasks

import (
	"neo3-squirrel/cache/address"
	"neo3-squirrel/cache/asset"
	"neo3-squirrel/db"
	"neo3-squirrel/rpc"
	"neo3-squirrel/tasks/applog"
	"neo3-squirrel/tasks/block"
	"neo3-squirrel/tasks/contract"
	"neo3-squirrel/tasks/nep17"
	"neo3-squirrel/util/log"
)

// Run manages all sync tasks.
func Run() {
	log.Info("Start Neo3 blockchain data parser.")

	// Cache all known addresses from DB.
	address.CacheMulti(db.GetAllAddresses())

	// Load all known assets from DB.
	assets := db.GetAllAssets()
	asset.UpdateMulti(assets)

	rpc.TraceBestHeight()

	block.StartBlockSyncTask()
	contract.StartContractTask()
	applog.StartApplicationLogSyncTask()
	nep17.StartNEP17TransferSyncTask()
}
