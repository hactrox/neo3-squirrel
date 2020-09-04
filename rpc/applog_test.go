package rpc

import (
	"fmt"
	"neo3-squirrel/tests"
	"testing"
)

func TestGetApplicationLog(t *testing.T) {
	setRPCforTest(tests.GetTestRPC())

	// Get transactions of block index 0.
	block := SyncBlock(0)
	if block == nil {
		t.Fatal("failed to get block of index 0")
	}

	for _, tx := range block.Tx {
		appLogResult := GetApplicationLog(0, tx.Hash)
		fmt.Println(appLogResult)
	}
}
