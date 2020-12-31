package applog

import (
	"neo3-squirrel/db"
	"neo3-squirrel/models"
)

func persistApplicationLogs(appLogChan <-chan *appLogInfo) {
	for result := range appLogChan {
		logResult := result.appLog
		blockIndex := result.BlockIndex
		blockTime := result.BlockTime

		notis := models.ParseApplicationLog(blockIndex, blockTime, logResult)
		if len(notis) == 0 {
			continue
		}

		// Persist contract management notificatoins.
		csNotis := []*models.Notification{}
		for _, noti := range notis {
			if noti.Contract == models.ManagementContract {
				csNotis = append(csNotis, noti)
			}
		}

		db.InsertAppLogNotifications(notis, csNotis)

		LastAppLogPK = notis[len(notis)-1].ID
	}
}
