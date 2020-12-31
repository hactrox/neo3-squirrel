package applog

import (
	"neo3-squirrel/rpc"
)

func queryAppLog(workers int, preAppLogChan <-chan *preAppLog) {
	for i := 0; i < workers; i++ {
		go func(ch <-chan *preAppLog) {
			for pre := range ch {
				appLogQueryResult := rpc.GetApplicationLog(pre.BlockIndex, pre.Hash)
				appLogs.Store(pre.Hash, appLogQueryResult)
			}
		}(preAppLogChan)
	}
}
