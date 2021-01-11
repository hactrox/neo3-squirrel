package block

import (
	"fmt"
	"neo3-squirrel/cache/block"
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
const bufferSize = 20000

var (
	// bestRPCHeight util.SafeCounter.
	prog         = progress.New()
	buffer       Buffer
	worker       Worker
	blockChannel chan *rpc.Block

	// bestBlockIndex traces the highest node height.
	// The value won't decrease even if node resync.
	bestBlockIndex int
)

// StartBlockSyncTask starts block sync tasks.
func StartBlockSyncTask() {
	lastBlockHeight := db.GetLastBlockHeight()
	bestBlockIndex := rpc.GetBestHeight()

	if lastBlockHeight == bestBlockIndex {
		prog.Finished = true
	}

	log.Info(color.Greenf("Block sync progress: %d/%d", lastBlockHeight, bestBlockIndex),
		color.BGreenf(", %d", bestBlockIndex-lastBlockHeight),
		color.Green(" blocks behind"))

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
		log.Info(color.Greenf("%s. Remaining workers=%d", hint, worker.num()))
	}()

	for {
		// Control size of the buffer.
		if buffer.Size() > bufferSize {
			time.Sleep(time.Millisecond * 20)
			continue
		}

		// Trace the one-way increment rpc best height.
		rpcBestHeight := rpc.GetBestHeight()
		if rpcBestHeight > bestBlockIndex {
			bestBlockIndex = rpcBestHeight
		}

		// Wait till any upstream fullnode is alive.
		for rpc.AllFullnodesDown() {
			time.Sleep(100 * time.Millisecond)
		}

		// Waiting for the next block.
		if nextHeight >= bestBlockIndex+1 &&
			(config.GetWorkers() == 1 || worker.num() == 1) {
			waiting(&waited, nextHeight)
			continue
		}

		// Get new block from upstream fullnodes.
		b := rpc.SyncBlock(uint(nextHeight))
		if b == nil {
			// Quit extra goroutines if beyond the latest block.
			if nextHeight >= bestBlockIndex &&
				!rpc.AllFullnodesDown() &&
				worker.shouldQuit() {
				return
			}

			nextHeight = buffer.GetHighest() + 1
			time.Sleep(1 * time.Second)
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

	if prog.Finished {
		*waited++
	}

	msg := fmt.Sprintf("Waiting for block index %d", nextHeight)
	msg += fmt.Sprintf(" (%s)", timeutil.ParseSeconds(uint64(*waited)))

	rpcBestHeight := rpc.GetBestHeight()
	if rpcBestHeight < nextHeight-1 && rpcBestHeight != -1 {
		lag := nextHeight - rpcBestHeight
		msg += color.Red(fmt.Sprintf(" (rpcBestHeight=%d, %d blocks behind)", rpcBestHeight, lag))
	}

	if (prog.Finished || rpcBestHeight < nextHeight-1) && !rpc.AllFullnodesDown() {
		log.Infof(msg)
	}
}

func arrangeBlock(dbHeight int, queue chan<- *rpc.Block) {
	const sleepTime = 20
	height := uint(dbHeight + 1)
	delay := 0

	for {
		for rpc.AllFullnodesDown() {
			time.Sleep(100 * time.Millisecond)
		}

		if b, ok := buffer.Pop(int(height)); ok {
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

		if delay >= 3000 && delay%1000 == 0 {
			log.Infof("Waited for %d seconds for block height [%d] in [arrangeBlock]\n", delay/1000, height)
		}

		if delay > 3000 && (delay-3000)%1000 == 0 {
			log.Warn(color.Yellowf("block height %d is missing while downloading blocks", height))
			getMissingBlock(height)
		}
	}
}

func getMissingBlock(height uint) {
	log.Infof("Try fetching given block of height: %d\n", height)

	b := rpc.SyncBlock(height)
	if b != nil {
		buffer.Put(b)
	}
}

func storeBlock(ch <-chan *rpc.Block) {
	var pendingBlockSize = 0
	rawBlocks := []*rpc.Block{}

	for block := range ch {
		rawBlocks = append(rawBlocks, block)
		pendingBlockSize += block.Size
		if pendingBlockSize >= 2*1024*1024 ||
			int(block.Index) == buffer.GetHighest() {
			store(rawBlocks)
			rawBlocks = nil
			pendingBlockSize = 0
		}
	}
}

var bestHeight int

func store(rawBlocks []*rpc.Block) {
	maxIndex := int(rawBlocks[len(rawBlocks)-1].Index)
	blocks := models.ParseBlocks(rawBlocks)
	txBulk := models.ParseTxs(rawBlocks)

	// Cache blocks to help other modules getting blocks quicker if cache hit.
	block.CacheBlocks(blocks)
	db.InsertBlock(blocks, txBulk)

	rpcBestHeight := rpc.GetBestHeight()
	if rpcBestHeight > bestHeight {
		bestHeight = rpcBestHeight
	}

	showBlockStorageProgress(int64(maxIndex), int64(bestHeight))
}
