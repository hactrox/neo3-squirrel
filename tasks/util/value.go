package util

import (
	"encoding/base64"
	"fmt"
	"math"
	"math/big"
	"neo3-squirrel/models"
	"neo3-squirrel/util/log"
	"strconv"
)

// GetReadableAmount returns decimals-formatted amount.
// E.g., 100000000 unit of GAS with 8 decimals will return 1.
func GetReadableAmount(amount, decimals *big.Float) *big.Float {
	dec, accuracy := decimals.Int64()
	if accuracy != big.Exact {

		err := fmt.Errorf("decimals convert from *big.Float to int64 not accurate")
		log.Panic(err)
	}

	decimalsFactor := big.NewFloat(math.Pow10(int(dec)))
	readableAmount := new(big.Float).Quo(amount, decimalsFactor)

	return readableAmount
}

func extractAddress(stackItem models.StackItem) (string, bool) {
	typ := stackItem.Type
	value := stackItem.Value

	switch typ {
	case "Any":
		return "", true
	case "ByteString":
		return ExtractAddressFromByteString(value.(string))
	default:
		err := fmt.Errorf("failed to parse address in type %s and value=%v", typ, value)
		log.Error(err)
		return "", false
	}
}

func extractString(stackItem models.StackItem) (string, bool) {
	typ := stackItem.Type
	value := stackItem.Value

	switch typ {
	case "Any":
		return "", true
	case "ByteString":
		bytes, err := base64.StdEncoding.DecodeString(value.(string))
		if err != nil {
			log.Error("Failed to extract value(%v) of 'ByteString' type", value)
			return "", false
		}

		return string(bytes), true
	default:
		log.Errorf("Unsupported string extract type: %s, value=%v", typ, value)
		return "", false
	}
}

func extractValue(stackItem models.StackItem) (*big.Float, bool) {
	typ := stackItem.Type
	value := stackItem.Value

	switch typ {
	case "Boolean":
		return big.NewFloat(0), true
	case "Integer":
		valStr := value.(string)
		val, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			log.Error("Failed to extract value(%v) of 'Integer' type", value)
			return nil, false
		}
		return new(big.Float).SetInt64(val), true
	default:
		return nil, false
	}
}
