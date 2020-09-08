package util

import (
	"encoding/hex"
	"neo3-squirrel/util/base58"
	"neo3-squirrel/util/byteutil"
)

// GetAddrScriptHash returns script hash of an address.
// E.g., NTdkuNTx38tQk3a5rnV9HPT96zqFHCb97h -> 0xb3f1f587042a20dd0eef2e47f137504f1419b054
func GetAddrScriptHash(address string) string {
	bytes, err := base58.CheckDecode(address)
	if err != nil {
		panic(err)
	}

	bytes = bytes[1:21]
	return hex.EncodeToString(byteutil.ReverseBytes(bytes))
}
