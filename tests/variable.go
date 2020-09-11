package tests

import (
	"os"
	"testing"
)

// GetTestRPC gets test fullnode rpc address from environment variable.
func GetTestRPC(t *testing.T) string {
	testRPC := os.Getenv("NEO3_SQUIRREL_TEST_RPC")
	if testRPC == "" {
		t.Fatal("'NEO3_SQUIRREL_TEST_RPC' must be set before rpc tests")
	}

	return testRPC
}
