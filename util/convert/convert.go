package convert

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strings"
)

const decimalPrecision = 128

// UInt64sToList converts uint64 array to string.
func UInt64sToList(ids []uint64) string {
	return intsToList(fmt.Sprint(ids))
}

// UInt16sToList converts uint16 array to string.
func UInt16sToList(ids []uint16) string {
	return intsToList(fmt.Sprint(ids))
}

// Float64ToDecimal converts float64 to big.Float.
func Float64ToDecimal(val float64) *big.Float {
	return ToDecimal(fmt.Sprintf("%.8f", val))
}

// DecimalNeg returns val with its sign negated.
func DecimalNeg(val *big.Float) *big.Float {
	val = val.SetPrec(decimalPrecision)
	return new(big.Float).Neg(val)
}

// ToDecimal returns *big.Float format of given decimal string,
// will return nil if input string is empty.
func ToDecimal(valueStr string) *big.Float {
	value, _ := new(big.Float).SetPrec(decimalPrecision).SetString(valueStr)
	return value
}

// AmountReadable returns decimals-formatted amount.
// E.g., 100000000 unit of GAS with 8 decimals will return 1.
func AmountReadable(amount *big.Float, decimals uint) *big.Float {
	decimalsFactor := big.NewFloat(math.Pow10(int(decimals)))
	readableAmount := new(big.Float).Quo(amount, decimalsFactor)

	return readableAmount
}

// BigFloatToString converts big.Float to string.
func BigFloatToString(value *big.Float) string {
	if value == nil {
		return ""
	}

	valueStr := value.Text('f', 18)
	valueStr = strings.TrimRight(valueStr, "0")
	valueStr = strings.TrimSuffix(valueStr, ".")

	return valueStr
}

// BytesToBigFloat converts byte array to *big.Float.
func BytesToBigFloat(data []byte) *big.Float {
	dataRev := ReverseBytes(data)
	val := new(big.Float).SetPrec(decimalPrecision).SetInt(new(big.Int).SetBytes(dataRev))
	return val
}

// BytesToBigInt converts byte array to *big.Int.
func BytesToBigInt(data []byte) *big.Int {
	dataRev := ReverseBytes(data)
	val := new(big.Int).SetBytes(dataRev)
	return val
}

// ByteStrToStr converts byte-string to string.
func ByteStrToStr(byteStr string) (string, error) {
	bytes, err := hex.DecodeString(byteStr)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// ReverseBytes reverses the given bytes,
// the origin bytes remain unchanged.
func ReverseBytes(raw []byte) []byte {
	reversed := make([]byte, len(raw))

	for i := len(raw) - 1; i >= 0; i-- {
		reversed[len(raw)-i-1] = raw[i]
	}

	return reversed
}

// ParseAmountStr transfers amount string to *big.Float,
// the returning result is (value, decimals, is convertion success)
func ParseAmountStr(amountStr string) (*big.Float, int, bool) {
	if len(amountStr) == 0 {
		return nil, 0, false
	}

	dotIndex := strings.Index(amountStr, ".")
	decimals := 0

	if dotIndex > 0 {
		amountStr = strings.TrimRight(amountStr, "0")
		decimals = len(amountStr) - 1 - dotIndex
	}

	amount, ok := new(big.Float).SetPrec(decimalPrecision).SetString(amountStr)

	return amount, decimals, ok
}

func intsToList(str string) string {
	return strings.Trim(strings.Replace(str, " ", ",", -1), "[]")
}
