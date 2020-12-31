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
	"strings"
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

	msg := ""

	if prog.Percentage.Cmp(big.NewFloat(100)) >= 0 {
		msg += color.Greenf("Block storage progress: %d. Block is fully synchronized.", maxIndex)
		// msg += "100%"
	} else {
		msg += fmt.Sprintf("Block storage progress: %d/%d, ", maxIndex, highestIndex)
		msg += fmt.Sprintf("%.4f%%", prog.Percentage)
	}

	if rpc.AllFullnodesDown() {
		msg = color.LightPurple("(sync from local buffer)") + msg
		prog.LastOutputTime = now
	} else {
		msg = fmt.Sprintf("%s%s", prog.RemainingTimeStr, msg)
	}

	msgs := []string{msg}
	if prog.Finished && !rpc.AllFullnodesDown() {
		msgs = append(msgs, appLogSyncProgressIndicator())
		msgs = append(msgs, nep5SyncProgressIndicator())
	}

	log.Infof(strings.Join(msgs, " "))
	prog.LastOutputTime = now
}

func appLogSyncProgressIndicator() string {
	LastAppLogPK := applog.LastAppLogPK
	lastNoti := db.GetLastNotification()

	if lastNoti == nil {
		return ""
	}

	offset := lastNoti.ID - LastAppLogPK
	if LastAppLogPK == 0 || offset == 0 {
		return ""
	}

	return fmt.Sprintf("[notificatoins left %d records]", offset)
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

	return fmt.Sprintf("[nep5 tx left %d records]", offset)
}
