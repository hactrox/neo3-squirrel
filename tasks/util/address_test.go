package util

import (
	"encoding/hex"
	"testing"
)

func TestGetAddressFromPublicKeyBytes(t *testing.T) {
	testCases := map[string]string{
		"02562e7ff2f939d160a7db692fb6edd5a2e6f85d6c92027f88b6604b867713a159": "NdhRqndwajjqzrdRTgTMiXdZzLBqSeRKR5",
		"02208aea0068c429a03316e37be0e3e8e21e6cda5442df4c5914a19b3a9b6de375": "NUnLWXALK2G6gYa7RadPLRiQYunZHnncxg",
	}

	for pubKey, addr := range testCases {
		bytes, err := hex.DecodeString(pubKey)
		if err != nil {
			t.Fatal(err)
		}

		get := GetAddressFromPublicKeyBytes(bytes)
		if addr != get {
			t.Fatalf("Failed to get address from public key, get=%s, want=%s", get, addr)
		}
	}
}
