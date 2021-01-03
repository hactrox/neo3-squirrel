package nep17

import (
	"math/big"
	"neo3-squirrel/cache/gas"
	"neo3-squirrel/models"
	"neo3-squirrel/rpc"
	"neo3-squirrel/tasks/util"
)

func updateGASTotalSupply(transfers []*models.Transfer) *big.Float {
	// Check if has GAS claim transfer.
	hasGASClaimTransfer := false

	for _, transfer := range transfers {
		if transfer.IsGASClaimTransfer() {
			hasGASClaimTransfer = true
			break
		}
	}

	bestBlock := rpc.GetBestHeight()
	if !hasGASClaimTransfer || int(gas.CachedTillBlockIndex()) >= bestBlock || bestBlock < 0 {
		return nil
	}

	gasTotalSupply, ok := util.QueryAssetTotalSupply(uint(bestBlock), models.GAS, 8)
	if !ok {
		return nil
	}

	gas.CacheGASTotalSupply(uint(bestBlock), gasTotalSupply)
	return gasTotalSupply
}
