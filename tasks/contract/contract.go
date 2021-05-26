package contract

import (
	"neo3-squirrel/db"
	"neo3-squirrel/models"
	"neo3-squirrel/rpc"
	"neo3-squirrel/tasks/util"
	"neo3-squirrel/util/color"
	"neo3-squirrel/util/log"
	"time"
)

// StartContractTask starts contract related tasks.
func StartContractTask() {
	log.Info(color.Green("Contract state sync task started"))
	lastContract := db.GetLastContract()

	// Insert native contracts if zero contracts.
	if lastContract == nil {
		persistNativeContracts()
	}

	go handleContractNotification()
}

func handleContractNotification() {
	nextCSNotiPK := db.GetContractNotiPK() + 1

	for {
		csNotis := db.GetContractNotifications(nextCSNotiPK, 100)
		if len(csNotis) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		for _, csNoti := range csNotis {
			notiApplied := handleCsNoti(csNoti)
			if !notiApplied {
				db.UpdateContractNotiPK(csNoti.ID)
			}
		}

		nextCSNotiPK = csNotis[len(csNotis)-1].ID + 1
	}
}

func handleCsNoti(csNoti *models.Notification) bool {
	if util.VMStateFault(csNoti.VMState) {
		return false
	}

	contractHash, ok := util.GetContractHash(csNoti)
	if !ok {
		return false
	}

	// Attention:
	// `rawContractState` can be nil if it was deleted already,
	// or won't be nil if contract A destroyed and
	// then redeploy again(got the same contract hash).
	rawContractState := rpc.GetContractState(csNoti.BlockIndex, contractHash)

	// Get sender from transaction detail.
	tx := db.GetTransaction(csNoti.Hash)
	if tx == nil {
		log.Panicf("Failed to get transaction detail of txid=%s", csNoti.Hash)
	}

	contractState := models.ParseContractState(
		csNoti.BlockIndex,
		csNoti.BlockTime,
		tx.Sender,
		csNoti.Hash,
		rawContractState,
	)

	switch models.EventName(csNoti.EventName) {
	case models.ContractDeployEvent:
		insertContract(contractState, csNoti, contractHash)
	case models.ContractUpdateEvent:
		updateContract(contractState, csNoti.ID, contractHash)
	case models.ContractDestroyEvent:
		deleteContract(contractHash, csNoti.ID)
	default:
		log.Panicf("Unsupported contract notification eventname: %s", csNoti.EventName)
	}

	showContractDBState(csNoti.BlockIndex, csNoti.BlockTime, contractHash, csNoti.EventName, contractState)

	return true
}

func insertContract(contractState *models.ContractState, csNoti *models.Notification, contractHash string) {
	contractState.State = string(models.ContractDeployEvent)

	// If this contract matches NEP-17 token standard.
	var nep17 *models.Asset
	if supportNEP17(contractState) {
		// `nep17` can be nil if failed to get info.
		nep17 = util.QueryNEP17AssetInfo(csNoti, contractHash)
	}

	db.InsertContract(contractState, csNoti.ID, contractHash, nep17)
}

func updateContract(contractState *models.ContractState, csNotiID uint, contractHash string) {
	contractState.State = string(models.ContractUpdateEvent)
	db.UpdateContract(contractState, csNotiID, contractHash)
}

func deleteContract(contractHash string, csNotiID uint) {
	db.DeleteContract(contractHash, csNotiID)
}

func supportNEP17(contractState *models.ContractState) bool {
	abi := contractState.Manifest.ABI

	abiMethods := map[string]bool{}

	for _, method := range abi.Methods {
		abiMethods[method.Name] = true
	}

	// https://github.com/neo-project/proposals/blob/a308ae3eabc240e2e837ababdd2314d0f0587bd8/nep-17.mediawiki
	requiredMethods := []string{
		"symbol",
		"decimals",
		"totalSupply",
		"balanceOf",
		"transfer",
	}

	for _, requiredMethod := range requiredMethods {
		if !abiMethods[requiredMethod] {
			return false
		}
	}

	return true
}
