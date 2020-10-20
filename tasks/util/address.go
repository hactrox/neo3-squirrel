package util

import (
	"encoding/base64"
	"encoding/hex"
	"neo3-squirrel/util/base58"
	"neo3-squirrel/util/byteutil"
	"neo3-squirrel/util/hashutil"
)

// GetAddrScriptHash returns script hash of an address.
// E.g., NTdkuNTx38tQk3a5rnV9HPT96zqFHCb97h -> b3f1f587042a20dd0eef2e47f137504f1419b054
func GetAddrScriptHash(address string) string {
	bytes, err := base58.CheckDecode(address)
	if err != nil {
		panic(err)
	}

	bytes = bytes[1:21]
	return hex.EncodeToString(byteutil.ReverseBytes(bytes))
}

// ExtractAddressFromByteString converts Neo address from byte string.
func ExtractAddressFromByteString(byteString string) (string, bool) {
	bytes, err := base64.StdEncoding.DecodeString(byteString)
	if err != nil {
		return "", false
	}

	bytes = append([]byte{0x35}, bytes...)
	return base58.CheckEncode(bytes), true
}

// GetAddressFromPublicKeyBase64 calcluates address from base64 encoded public key.
func GetAddressFromPublicKeyBase64(base64Str string) (string, bool) {
	bytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return "", false
	}

	return GetAddressFromPublicKeyBytes(bytes), true
}

// GetAddressFromPublicKeyBytes calculates address from public key hex string.
func GetAddressFromPublicKeyBytes(bytes []byte) string {
	bytes = append([]byte{0x0C, 0x21}, bytes...)
	bytes = append(bytes, 0x0b, 0x41, 0x95, 0x44, 0x0d, 0x78)
	bytes = append([]byte{0x35}, hashutil.Hash160(bytes)...)

	return base58.CheckEncode(bytes)
}
