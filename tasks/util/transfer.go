package util

import (
	"math/big"
	"neo3-squirrel/models"
)

// ExtractNEP5Transfer extracts NEP5 transfer stack items into readable variables.
func ExtractNEP5Transfer(stackItems []models.StackItem) (from string, to string, amount *big.Float, ok bool) {
	from, ok = extractAddress(stackItems[0].Type, stackItems[0].Value)
	if !ok {
		return
	}
	to, ok = extractAddress(stackItems[1].Type, stackItems[1].Value)
	if !ok {
		return
	}
	amount, ok = extractValue(stackItems[2].Type, stackItems[2].Value)
	if !ok {
		return
	}

	return
}
