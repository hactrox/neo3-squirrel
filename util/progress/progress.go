package progress

import (
	"fmt"
	"math"
	"math/big"
	"neo3-squirrel/util/timeutil"
	"time"
)

// Progress stores progress info of a task.
type Progress struct {
	InitPercentage   *big.Float
	InitTime         time.Time
	Percentage       *big.Float
	RemainingTimeStr string
	// Finished indicates if fully synced(current task).
	Finished bool
	// MailSent is a mark that when fully synced, send notify mail once.
	MailSent       bool
	LastOutputTime time.Time
}

func (progInfo *Progress) updatePercentage(percentage *big.Float) {
	percentage = new(big.Float).Mul(percentage, big.NewFloat(10000))
	val, _ := percentage.Int64()
	progInfo.Percentage = new(big.Float).SetFloat64(float64(val) / 10000)
}

func (progInfo *Progress) extractSeconds(secondsLeft uint64) {
	// It is meaningless to show remaining time after block height is up to date or a task had finished.
	if progInfo.Finished {
		progInfo.RemainingTimeStr = ""
	} else {
		timeStr := timeutil.ParseSeconds(secondsLeft)
		progInfo.RemainingTimeStr = fmt.Sprintf("(%s left)", timeStr)
	}
}

// GetEstimatedRemainingTime calculates remaining time of a task.
func GetEstimatedRemainingTime(curr int64, total int64, progInfo *Progress) {
	percentage := new(big.Float).Quo(new(big.Float).SetInt64(curr), new(big.Float).SetInt64(total))
	percentage = new(big.Float).Mul(percentage, big.NewFloat(100))

	if (*progInfo).InitTime == (time.Time{}) {
		progInfo.InitPercentage = percentage
		progInfo.InitTime = time.Now()
		progInfo.updatePercentage(percentage)
		return
	}

	if curr >= total {
		// progInfo.Finished = true.
		progInfo.extractSeconds(0)
		progInfo.Percentage = big.NewFloat(100)
		return
	}

	elapsedTime := time.Since(progInfo.InitTime)
	progInfo.updatePercentage(percentage)

	elaspedPercentage := new(big.Float).Sub(percentage, progInfo.InitPercentage)
	// Incase denominator increases more rapidly, e.g. 10/100 becomes 11/110.
	if elaspedPercentage.Cmp(big.NewFloat(0)) < 0 {
		progInfo.InitPercentage = percentage
		progInfo.InitTime = time.Now()
		progInfo.Finished = false
		return
	}

	estimatedRemainingTime := new(big.Float).Quo(new(big.Float).SetFloat64(elapsedTime.Seconds()), elaspedPercentage)
	estimatedRemainingTime = new(big.Float).Mul(estimatedRemainingTime, new(big.Float).Sub(big.NewFloat(100), percentage))
	secondsLeft, _ := estimatedRemainingTime.Float64()
	progInfo.extractSeconds(uint64(math.Ceil(secondsLeft)))
}
