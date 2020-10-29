package nep5

import (
	"fmt"
	"neo3-squirrel/cache/asset"
	"neo3-squirrel/models"
	"neo3-squirrel/util/color"
	"neo3-squirrel/util/convert"
	"neo3-squirrel/util/log"
)

func showTransfers(transfers []*models.Transfer) {
	for _, transfer := range transfers {
		from := transfer.From
		to := transfer.To
		amount := transfer.Amount
		contractHash := transfer.Contract
		contract, ok := asset.GetNEP5(contractHash)
		if !ok {
			log.Panicf("Failed to get asset info of contract %s", contractHash)
		}

		msg := ""
		symbol := contract.Symbol
		amountWithUnit := fmt.Sprintf("%s %s", convert.BigFloatToString(amount), symbol)

		if len(from) == 0 {
			// Claim GAS.
			if contractHash == models.GAS {
				msg = color.Greenf("   GAS claimed: %s + %s", to, amountWithUnit)
			} else {
				msg = color.LightGreenf("  Token minted: %s + %s", to, amountWithUnit)
			}
		} else {
			if len(to) == 0 {
				msg = color.LightPurplef(" Destroy token: %s - %s", from, amountWithUnit)
			} else {
				msg = color.LightCyanf("Token transfer: %s -> %s, amount %s", from, to, amountWithUnit)
			}
		}

		log.Info(msg)
	}
}
