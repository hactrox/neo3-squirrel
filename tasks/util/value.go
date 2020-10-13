package util

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"neo3-squirrel/models"
	"neo3-squirrel/util/convert"
	"neo3-squirrel/util/log"
	"strconv"
)

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
	case "Integer":
		return value.(string), true
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
		boolVal := value.(bool)
		if boolVal {
			return convert.BigFloatFromInt64(1), true
		}

		return convert.BigFloatFromInt64(0), true
	case "Integer":
		valStr := value.(string)
		val, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			log.Error("Failed to extract value(%v) of 'Integer' type", value)
			return nil, false
		}

		return convert.BigFloatFromInt64(val), true
	default:
		return nil, false
	}
}
