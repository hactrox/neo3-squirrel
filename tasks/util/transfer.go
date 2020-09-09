package util

import (
	"math/big"
	"neo3-squirrel/models"
)

// ExtractNEP5Transfer extracts NEP5 transfer stack items into readable variables.
func ExtractNEP5Transfer(stackItems []models.StackItem) (from string, to string, amount *big.Float, ok bool) {
	from, ok = extractAddress(stackItems[0])
	if !ok {
		return
	}
	to, ok = extractAddress(stackItems[1])
	if !ok {
		return
	}
	amount, ok = extractValue(stackItems[2])
	return
}
