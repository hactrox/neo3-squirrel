package applog

import (
	"neo3-squirrel/rpc"
)

func queryAppLog(workers int, preAppLogChan <-chan *preAppLog) {
	for i := 0; i < workers; i++ {
		go func(ch <-chan *preAppLog) {
			for pre := range ch {
				// log.Debugf("prepare to query applicationlog of tx %s", tx.Hash)
				appLogQueryResult := rpc.GetApplicationLog(pre.BlockIndex, pre.Hash)
				appLogs.Store(pre.Hash, appLogQueryResult)
				// log.Debugf("store applog result into appLogs, txID=%s, len(noti)=%d", tx.Hash, len(appLogQueryResult.Notifications))
			}
		}(preAppLogChan)
	}
}
