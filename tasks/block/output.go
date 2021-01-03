package block

import (
	"fmt"
	"math/big"
	"neo3-squirrel/db"
	"neo3-squirrel/rpc"
	"neo3-squirrel/tasks/applog"
	"neo3-squirrel/tasks/nep17"
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
		msgs = append(msgs, appLogSyncProgressIndicator(uint(maxIndex)))
		msgs = append(msgs, nep17SyncProgressIndicator(uint(maxIndex)))
	}

	log.Infof(strings.Join(msgs, " "))
	prog.LastOutputTime = now
}

func appLogSyncProgressIndicator(currBlockIndex uint) string {
	lastBlockIndex := applog.LastAppLogBlockIndex
	lastNoti := db.GetLastNotification()

	if lastNoti == nil {
		return ""
	}

	// If only 1 block behind.
	if lastBlockIndex >= currBlockIndex-1 {
		return ""
	}

	offset := lastNoti.BlockIndex - lastBlockIndex
	if lastBlockIndex == 0 || offset == 0 {
		return ""
	}

	return fmt.Sprintf("[notificatoins left %d blocks]", offset)
}

func nep17SyncProgressIndicator(currBlockIndex uint) string {
	lastBlockIndex := nep17.LastTxBlockIndex
	lastNoti := db.GetLastNotiForNEP17Task()

	if lastNoti == nil {
		return ""
	}

	// If only 1 block behind.
	if lastBlockIndex >= currBlockIndex-1 {
		return ""
	}

	offset := lastNoti.BlockIndex - lastBlockIndex
	if lastBlockIndex == 0 || offset == 0 {
		return ""
	}

	return fmt.Sprintf("[nep17 tx left %d blocks]", offset)
}
