package applog

import (
	"neo3-squirrel/cache/contractstate"
	"neo3-squirrel/db"
	"neo3-squirrel/models"
	"neo3-squirrel/util/convert"
)

func persistApplicationLogs(appLogChan <-chan *appLogResult) {
	// TODO: mail alert

	for result := range appLogChan {
		logResult := result.appLogQueryResult
		blockIndex := result.BlockIndex
		blockTime := result.BlockTime

		// Handle contract add, migrate, delete actions.
		csList := contractstate.PopFirstIf(blockIndex)
		if len(csList) > 0 {
			handleContractStateChange(csList)
		}

		// log.Debugf("handle application log of txID: %s", result.Hash)

		// Store applicationlog result
		appLog := models.ParseApplicationLog(blockIndex, blockTime, logResult)
		appLog.GasConsumed = convert.AmountReadable(appLog.GasConsumed, 8)
		db.InsertApplicationLog(appLog)
	}
}
