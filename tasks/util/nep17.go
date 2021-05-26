package util

import (
	"neo3-squirrel/cache/gas"
	"neo3-squirrel/config"
	"neo3-squirrel/models"
	"neo3-squirrel/rpc"
	"neo3-squirrel/util/log"
	"neo3-squirrel/util/timeutil"
)

func QueryNEP17AssetInfo(noti *models.Notification, contractHash string) *models.Asset {
	blockIndex := noti.BlockIndex

	// Get and save contract manifest.
	contractState := rpc.GetContractState(blockIndex, contractHash)
	if contractState == nil {
		return nil
	}

	asset := models.Asset{
		BlockIndex: blockIndex,
		BlockTime:  noti.BlockTime,
		TxID:       noti.Hash,
		Contract:   contractHash,
		Name:       contractState.Manifest.Name,
	}

	bestBlockIndex := rpc.GetBestHeight()
	ok := queryAssetBasicInfo(blockIndex, &asset)
	if !ok {
		log.Warnf("Failed to get NEP17 contract info. Hash=%s(%s), Contract=%s, BlockIndex=%d, BlockTime=%s",
			noti.Hash, noti.Src, contractHash, blockIndex, timeutil.FormatBlockTime(noti.BlockTime))
		return nil
	}

	if contractHash == models.GasToken && bestBlockIndex > 0 {
		gas.CacheGASTotalSupply(uint(bestBlockIndex), asset.TotalSupply)
	}

	if asset.TotalSupply != nil && asset.TotalSupply.Cmp(config.MaxVal) > 0 {
		log.Errorf("Asset total supply , asset skipped: %v", asset)
	}

	return &asset
}
