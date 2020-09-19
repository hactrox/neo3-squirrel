package base58

import (
	"bytes"
	"errors"
	"neo3-squirrel/util/hashutil"

	"github.com/mr-tron/base58"
)

// CheckDecode decides base58 with checksum check.
func CheckDecode(s string) (b []byte, err error) {
	b, err = base58.Decode(s)
	if err != nil {
		return nil, err
	}

	if len(b) < 5 {
		return nil, errors.New("invalid base-58 check string: missing checksum")
	}

	if !bytes.Equal(hashutil.Checksum(b[:len(b)-4]), b[len(b)-4:]) {
		return nil, errors.New("invalid base-58 check string: invalid checksum")
	}

	// Trim last 4 bytes.
	b = b[:len(b)-4]

	return b, nil
}

// CheckEncode encodes the given bytes into
// base58 encoding with checksum appended to it.
func CheckEncode(b []byte) string {
	b = append(b, hashutil.Checksum(b)...)
	return base58.Encode(b)
}
