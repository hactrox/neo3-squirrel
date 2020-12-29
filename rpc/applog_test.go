package rpc

import (
	"neo3-squirrel/tests"
	"neo3-squirrel/util/log"
	"os"
	"testing"
)

func TestGetApplicationLog(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	log.Init(true)
	setRPCforTest(tests.GetTestRPC(t))
	defer func() {
		os.RemoveAll("./logs")
	}()

	if bestHeight.Get() < 0 {
		t.Skip("No upstream fullnode available, test skipped")
	}

	// Get transactions of block index 0.
	block := SyncBlock(0)
	if block == nil {
		t.Fatal("failed to get block of index 0")
	}

	appLogResult := GetApplicationLog(0, block.Hash)
	if appLogResult == nil ||
		appLogResult.BlockHash == "" {
		t.Fatalf("Incorrect 'GetApplicationLog' func, txid=%s", block.Hash)
	}
}
