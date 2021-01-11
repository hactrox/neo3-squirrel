package contract

import (
	"fmt"
	"neo3-squirrel/models"
	"neo3-squirrel/util/color"
	"neo3-squirrel/util/log"
	"neo3-squirrel/util/timeutil"
)

func showContractDBState(cs *models.ContractState) {
	blockInfo := fmt.Sprintf("(block %d %s)", cs.BlockIndex, timeutil.FormatBlockTime(cs.BlockTime))
	msg := ""

	switch models.EventName(cs.State) {
	case models.ContractDeployEvent:
		msg = "Contract deployed:"
	case models.ContractUpdateEvent:
		msg = " Contract updated:"
	case models.ContractDestroyEvent:
		msg = " Contract deleted:"
	default:
		msg = "Unknown Contract state:"
	}

	msg += fmt.Sprintf(" %s %s %-17s", blockInfo, cs.Hash, cs.Manifest.Name)

	if len(cs.Manifest.SupportedStandards) > 0 {
		msg += fmt.Sprintf(", support=%v", cs.Manifest.SupportedStandards)
	}

	msg = color.BCyan(msg)
	log.Info(msg)
}
