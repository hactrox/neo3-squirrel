package block

import (
	"fmt"
	"math/big"
	"neo3-squirrel/config"
	"neo3-squirrel/db"
	"neo3-squirrel/models"
	"neo3-squirrel/rpc"
	"neo3-squirrel/util/color"
	"neo3-squirrel/util/log"
	"neo3-squirrel/util/progress"
	"neo3-squirrel/util/timeutil"
	"time"
)

// bufferSize is the capacity of pending blocks waiting to be persisted to db.
const bufferSize = 5000

var (
	// bestRPCHeight util.SafeCounter.
	prog         = progress.Progress{}
	buffer       Buffer
	worker       Worker
	blockChannel chan *rpc.Block
)

// StartBlockSyncTask start block sync tasks.
func StartBlockSyncTask() {
	lastBlockHeight := db.GetLastBlockHeight()
	buffer = NewBuffer(lastBlockHeight)

	for i := 0; i < config.GetWorkers(); i++ {
		go fetchBlock()
	}

	blockChannel = make(chan *rpc.Block, bufferSize)
	go arrangeBlock(lastBlockHeight, blockChannel)
	go storeBlock(blockChannel)
}

func fetchBlock() {
	worker.add()
	log.Infof("Create new worker to fetch blocks\n")

	nextHeight := buffer.GetNextPending()
	waited := 0

	defer func() {
		const hint = "Worker for block data persistence terminated"
		log.Info(color.BGreenf("%s. Remaining workers=%d", hint, worker.num()))
	}()

	// TODO: mail alert
	// defer mail.AlertIfErr()

	for {
		// Control size of the buffer.
		if buffer.Size() > bufferSize {
			time.Sleep(time.Millisecond * 20)
			continue
		}

		if config.GetWorkers() == 1 {
			if nextHeight == rpc.GetBestHeight()+1 {
				waiting(&waited, nextHeight)
			}
		} else if worker.num() == 1 && nextHeight == buffer.GetHighest()+1 {
			waiting(&waited, nextHeight)
		}

		b := rpc.SyncBlock(nextHeight)

		// Beyond the latest block.
		if b == nil {
			if nextHeight > rpc.GetBestHeight()-50 &&
				worker.shouldQuit() {
				return
			}

			// Get the correct next pending block.
			nextHeight = buffer.GetHighest() + 1
			continue
		}

		waited = 0
		buffer.Put(b)

		if worker.num() == 1 {
			nextHeight = buffer.GetHighest() + 1
		} else {
			nextHeight = buffer.GetNextPending()
		}
	}
}

func waiting(waited *int, nextHeight int) {
	time.Sleep(time.Second)
	*waited++
	log.Infof("Waiting for block index: %d(%s)\n", nextHeight, timeutil.ParseSeconds(uint64(*waited)))
	// if waited >= 30 && waited%10 == 0 {
	// 	rpc.SwitchServer()
	// }
}

func arrangeBlock(dbHeight int, queue chan<- *rpc.Block) {
	// TODO: mail alert
	// defer mail.AlertIfErr()

	const sleepTime = 20
	height := dbHeight + 1
	delay := 0

	for {
		if b, ok := buffer.Pop(height); ok {
			queue <- b
			height++
			delay = 0
			continue
		}

		time.Sleep(time.Millisecond * time.Duration(sleepTime))
		if buffer.Size() == 0 {
			continue
		}
		delay += sleepTime

		if delay >= 5000 && delay%1000 == 0 {
			log.Infof("Waited for %d seconds for block height [%d] in [arrangeBlock]\n", delay/1000, height)
		}

		if delay%(1000*10) == 0 {
			err := fmt.Errorf("block height %d is missing while downloading blocks", height)
			log.Warn(err)

			getMissingBlock(height)
		}
	}
}

func getMissingBlock(height int) {
	log.Infof("Try fetching given block of height: %d\n", height)

	b := rpc.SyncBlock(height)
	if b != nil {
		buffer.Put(b)
	}
}

func storeBlock(ch <-chan *rpc.Block) {
	// TODO: mail alert
	// defer mail.AlertIfErr()

	const size = 15
	rawBlocks := []*rpc.Block{}

	for block := range ch {
		rawBlocks = append(rawBlocks, block)
		if block.Index%size == 0 ||
			int(block.Index) == buffer.GetHighest() {
			store(rawBlocks)
			rawBlocks = nil
		}
	}
}

func store(rawBlocks []*rpc.Block) {
	maxIndex := int(rawBlocks[len(rawBlocks)-1].Index)
	blocks := models.ParseBlocks(rawBlocks)
	txBulk := models.ParseTxs(rawBlocks)

	db.InsertBlock(blocks, txBulk)

	// Auxiliary signal for tx task.
	// TxMaxPkShouldRefresh = true
	// AssetTxMaxPkShouldRefresh = true
	// Nep5MaxPkShouldRefresh = true
	// nftMaxPkShouldRefresh = true
	// gasMaxPkShouldRefresh = true
	// scMaxPkShouldRefresh = true

	bestHeight := rpc.GetBestHeight()

	if bestHeight < maxIndex {
		bestHeight = maxIndex
	}

	showBlockStorageProgress(int64(maxIndex), int64(bestHeight))
}

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

	log.Infof("%sBlock storage progress: %d/%d, %.4f%%\n",
		prog.RemainingTimeStr,
		maxIndex,
		highestIndex,
		prog.Percentage,
	)
	prog.LastOutputTime = now

	// Send mail if fully synced.
	if prog.Finished && !prog.MailSent {
		prog.MailSent = true

		// If sync lasts shortly, do not send mail.
		if time.Since(prog.InitTime) < time.Minute*5 {
			return
		}

		// TODO: mail alert
		// msg := fmt.Sprintf("Block counts: %d", highestIndex)
		// mail.SendNotify("Block data Fully Synced", msg)
	}
}