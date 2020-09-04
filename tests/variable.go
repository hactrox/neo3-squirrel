package tests

import (
	"neo3-squirrel/util/log"
	"os"
)

func init() {
	log.Init(true)
}

// GetTestRPC gets test fullnode rpc address from environment variable.
func GetTestRPC() string {
	return os.Getenv("NEO3_SQUIRREL_TEST_RPC")
}
