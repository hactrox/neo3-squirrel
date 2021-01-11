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

	rawContractState := rpc.GetContractState(csNoti.BlockIndex, contractHash)
	if rawContractState == nil {
		// This contract may be deleted so we cannot get contract state from rpc.
		return false
	}

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
		insertContract(contractState, csNoti.ID)
	case models.ContractUpdateEvent:
		updateContract(contractState, csNoti.ID, contractHash)
	case models.ContractDestroyEvent:
		deleteContract(contractState.ID, csNoti.ID)
	default:
		log.Panicf("Unsupported contract notification eventname: %s", csNoti.EventName)
	}

	showContractDBState(contractState)

	return true
}

func insertContract(contractState *models.ContractState, csNotiID uint) {
	contractState.State = string(models.ContractDeployEvent)
	db.InsertContract(contractState, csNotiID)
}

func updateContract(contractState *models.ContractState, csNotiID uint, contractHash string) {
	contractState.State = string(models.ContractUpdateEvent)
	db.UpdateContract(contractState, csNotiID, contractHash)
}

func deleteContract(csID uint, csNotiID uint) {
	db.DeleteContract(csID, csNotiID)
}
