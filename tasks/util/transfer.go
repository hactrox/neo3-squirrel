package util

import (
	"math/big"
	"neo3-squirrel/models"
)

// ExtractNEP17Transfer extracts NEP17 transfer stack items into readable variables.
func ExtractNEP17Transfer(stackItems []models.StackItem) (from string, to string, rawAmount *big.Float, ok bool) {
	from, ok = extractAddress(stackItems[0])
	if !ok {
		return
	}
	to, ok = extractAddress(stackItems[1])
	if !ok {
		return
	}
	rawAmount, ok = extractValue(stackItems[2])
	return
}
