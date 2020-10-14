package applog

import (
	"neo3-squirrel/db"
	"neo3-squirrel/models"
	"neo3-squirrel/util/convert"
)

func persistApplicationLogs(appLogChan <-chan *appLogResult) {
	for result := range appLogChan {
		logResult := result.appLogQueryResult
		blockIndex := result.BlockIndex
		blockTime := result.BlockTime

		// log.Debugf("handle application log of txID: %s", result.Hash)

		// Store applicationlog result
		appLog := models.ParseApplicationLog(blockIndex, blockTime, logResult)
		appLog.GasConsumed = convert.AmountReadable(appLog.GasConsumed, 8)
		db.InsertApplicationLog(appLog)
		LastTxPK = result.PK
	}
}
