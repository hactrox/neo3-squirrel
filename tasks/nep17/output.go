package nep17

import (
	"fmt"
	"neo3-squirrel/cache/asset"
	"neo3-squirrel/models"
	"neo3-squirrel/util/color"
	"neo3-squirrel/util/convert"
	"neo3-squirrel/util/log"
	"neo3-squirrel/util/timeutil"
)

func showTransfers(transfers []*models.Transfer) {
	for _, transfer := range transfers {
		from := transfer.From
		to := transfer.To
		amount := transfer.Amount
		contractHash := transfer.Contract
		contract, ok := asset.Get(contractHash)
		if !ok {
			log.Panicf("Failed to get asset info of contract %s", contractHash)
		}

		msg := ""
		symbol := contract.Symbol
		amountWithUnit := fmt.Sprintf("%s %s", convert.BigFloatToString(amount), symbol)

		blockInfo := fmt.Sprintf("(block %d %s)", transfer.BlockIndex, timeutil.FormatBlockTime(transfer.BlockTime))

		if len(from) == 0 {
			// Claim GAS.
			if contractHash == models.GASContract {
				content := fmt.Sprintf("%s System Reward: %s + %s", blockInfo, to, amountWithUnit)
				msg = color.Greenf(content)
			} else {
				content := fmt.Sprintf("%s  Token Minted: %s + %s", blockInfo, to, amountWithUnit)
				msg = color.LightGreenf(content)
			}
		} else {
			if len(to) == 0 {
				content := fmt.Sprintf("%s Token Destroy: %s - %s", blockInfo, from, amountWithUnit)
				msg = color.LightPurplef(content)
			} else {
				content := fmt.Sprintf("%sToken Transfer: %s -> %s: %s", blockInfo, from, to, amountWithUnit)
				msg = color.LightCyanf(content)
			}
		}

		log.Info(msg)
	}
}
