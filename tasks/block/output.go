package block

import (
	"fmt"
	"math/big"
	"neo3-squirrel/db"
	"neo3-squirrel/rpc"
	"neo3-squirrel/tasks/applog"
	"neo3-squirrel/tasks/nep5"
	"neo3-squirrel/util/color"
	"neo3-squirrel/util/log"
	"neo3-squirrel/util/progress"
	"time"
)

func showBlockStorageProgress(maxIndex int64, highestIndex int64) {
	now := time.Now()

	if prog.LastOutputTime == (time.Time{}) {
		prog.LastOutputTime = now
	}

	if maxIndex < highestIndex &&
		now.Sub(prog.LastOutputTime) < time.Second {
		return
	}

	progress.GetEstimatedRemainingTime(maxIndex, highestIndex, &prog)
	if prog.Percentage.Cmp(big.NewFloat(100)) == 0 {
		prog.Finished = true
	}

	msg := fmt.Sprintf("Block storage progress: %d/%d, ", maxIndex, highestIndex)

	if prog.Percentage.Cmp(big.NewFloat(100)) >= 0 {
		msg += "100%"
	} else {
		msg += fmt.Sprintf("%.4f%%", prog.Percentage)
	}

	if rpc.AllFullnodesDown() {
		msg = color.BLightPurple("(sync from local buffer)") + msg
		prog.LastOutputTime = now
	} else {
		msg = fmt.Sprintf("%s%s", prog.RemainingTimeStr, msg)
	}

	if prog.Finished && !rpc.AllFullnodesDown() {
		msg += appLogSyncProgressIndicator()
		msg += nep5SyncProgressIndicator()
	}

	log.Infof(msg)
	prog.LastOutputTime = now
}

func appLogSyncProgressIndicator() string {
	lastPersistedPK := applog.LastTxPK
	LastTx := db.GetLastTransaction()

	if LastTx == nil {
		return ""
	}

	offset := LastTx.ID - lastPersistedPK

	if lastPersistedPK == 0 || offset == 0 {
		return ""
	}

	return fmt.Sprintf(" [appLog left %d records]", offset)
}

func nep5SyncProgressIndicator() string {
	lastPersistedPK := nep5.LastTxPK
	lastNoti := db.GetLastNotiForNEP5Task()

	if lastNoti == nil {
		return ""
	}

	offset := lastNoti.ID - lastPersistedPK

	if lastPersistedPK == 0 || offset == 0 {
		return ""
	}

	return fmt.Sprintf(" [nep5 tx left %d records]", offset)
}
