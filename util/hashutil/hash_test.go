package hashutil

import (
	"fmt"
	"testing"
)

func TestHash160(t *testing.T) {
	data := []byte{0x00, 0x01, 0x02, 0x03}
	want := "3c3fa3d4adcaf8f52d5b1843975e122548269937"
	get := Hash160(data)

	if fmt.Sprintf("%x", get) != want {
		t.Fatalf("Get: %x, want: %s", get, want)
	}
}

func TestHash256(t *testing.T) {
	data := []byte{0x00, 0x01, 0x02, 0x03}
	want := "f7a355c00c89a08c80636bed35556a210b51786f6803a494f28fc5ba05959fc2"
	get := Hash256(data)

	if fmt.Sprintf("%x", get) != want {
		t.Fatalf("Get: %x, want: %s", get, want)
	}
}

func TestGetScriptHash(t *testing.T) {
	data := []byte{0x00, 0x01, 0x02, 0x03}
	want := "3c3fa3d4adcaf8f52d5b1843975e122548269937"
	get := GetScriptHash(data)

	if fmt.Sprintf("%x", get) != want {
		t.Fatalf("Get: %x, want: %s", get, want)
	}
}

func TestGetAssetIDFromScriptHash(t *testing.T) {
	data := []byte{0x00, 0x01, 0x02, 0x03}
	want := "3033303230313030"
	get := GetAssetIDFromScriptHash(data)

	if fmt.Sprintf("%x", get) != want {
		t.Fatalf("Get: %x, want: %s", get, want)
	}
}

func TestGetScriptHashFromAssetID(t *testing.T) {
	assetID := "43cf98eddbe047e198a3e5d57006311442a0ca15"
	want := "15caa04214310670d5e5a398e147e0dbed98cf43"
	get, err := GetScriptHashFromAssetID(assetID)
	if err != nil {
		t.Fatal(err)
	}

	if fmt.Sprintf("%x", get) != want {
		t.Fatalf("Get: %x, want: %s", get, want)
	}
}

func TestGetScriptHashFromAssetIDPanic(t *testing.T) {
	_, err := GetScriptHashFromAssetID("squirrel")
	if err == nil {
		t.Fatalf("Get error=nil, want an error")
	}
}
