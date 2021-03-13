package contract

import (
	"fmt"
	"neo3-squirrel/models"
	"neo3-squirrel/util/color"
	"neo3-squirrel/util/log"
	"neo3-squirrel/util/timeutil"
	"strings"
)

func showContractDBState(blockIndex uint, blockTime uint64, contractHash, eventName string, cs *models.ContractState) {
	blockInfo := fmt.Sprintf("(block %d %s)", blockIndex, timeutil.FormatBlockTime(blockTime))
	msg := ""

	switch models.EventName(eventName) {
	case models.ContractDeployEvent:
		msg = "Contract deployed:"
	case models.ContractUpdateEvent:
		msg = "Contract updated:"
	case models.ContractDestroyEvent:
		msg = "Contract destroyed:"
	default:
		msg = "Unknown Contract state:"
	}

	msg += fmt.Sprintf(" %s %s", blockInfo, contractHash)

	if cs != nil {
		if models.EventName(eventName) != models.ContractDestroyEvent {
			msg += fmt.Sprintf(" %s", cs.Manifest.Name)
		}

		if len(cs.Manifest.SupportedStandards) > 0 {
			msg += fmt.Sprintf(", support=[%s]", strings.Join(cs.Manifest.SupportedStandards, ", "))
		}
	}

	msg = color.BCyan(msg)
	log.Info(msg)
}
