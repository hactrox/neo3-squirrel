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

		symbol := contract.Symbol
		msg := ""
		amountStr := convert.BigFloatToString(amount)

		if len(from) == 0 {
			// Claim GAS.
			if contractHash == models.GAS {
				msg = fmt.Sprintf("   GAS claimed: %s get %s %s", to, amountStr, symbol)
				msg = color.Green(msg)
			} else {
				msg = fmt.Sprintf("  Token minted: %s get %s %s", to, amountStr, symbol)
				msg = color.LightGreen(msg)
			}
		} else {
			if len(to) == 0 {
				msg = fmt.Sprintf(" Destroy token: %s lost %s %s", from, amountStr, symbol)
				msg = color.LightPurple(msg)
			} else {
				msg = fmt.Sprintf("Token transfer: %s -> %s, amount %s %s", from, to, amountStr, symbol)
				msg = color.LightCyan(msg)
			}
		}

		log.Info(msg)
	}
}
