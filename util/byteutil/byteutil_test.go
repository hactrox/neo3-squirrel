package byteutil

import (
	"bytes"
	"testing"
)

func TestReverseBytes(t *testing.T) {
	rawBytes := []byte{0x01, 0x02, 0x03}
	want := []byte{0x03, 0x02, 0x01}
	get := ReverseBytes(rawBytes)
	if !bytes.Equal(get, want) {
		t.Fatalf("Get=%v, want=%v", get, want)
	}

	rawBytes = nil
	want = nil
	get = ReverseBytes(rawBytes)
	if !bytes.Equal(get, want) {
		t.Fatalf("Get=%v, want=%v", get, want)
	}

	rawBytes = []byte{0x00}
	want = []byte{0x00}
	get = ReverseBytes(rawBytes)
	if !bytes.Equal(get, want) {
		t.Fatalf("Get=%v, want=%v", get, want)
	}

	rawBytes = []byte{0x00, 0x01}
	want = []byte{0x01, 0x00}
	get = ReverseBytes(rawBytes)
	if !bytes.Equal(get, want) {
		t.Fatalf("Get=%v, want=%v", get, want)
	}
}
