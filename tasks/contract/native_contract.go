package contract

import (
	"neo3-squirrel/db"
	"neo3-squirrel/models"
	"neo3-squirrel/rpc"
	"neo3-squirrel/util/log"
)

func persistNativeContracts() {
	nativeContractHashes := models.NativeContractHashes()

	genesisBlock := rpc.SyncBlock(0)
	if genesisBlock == nil {
		log.Panicf("Failed to get genesis block from Fullnode RPC")
	}

	blockIndex := genesisBlock.Index
	blockTime := genesisBlock.Time

	for _, contractHash := range nativeContractHashes {
		RawContractState := rpc.GetContractState(0, contractHash)
		if RawContractState == nil {
			log.Panicf("Failed to get contract state of hash %s", contractHash)
		}

		cs := models.ParseContractState(blockIndex, blockTime, "", "", RawContractState)
		db.InsertNativeContract(cs)

		showContractDBState(
			blockIndex,
			blockTime,
			contractHash,
			string(models.ContractDeployEvent), cs,
		)
	}
}
