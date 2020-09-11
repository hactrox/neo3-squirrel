package nep5

import (
	"math/big"
	"neo3-squirrel/models"
	"neo3-squirrel/rpc"
	"neo3-squirrel/tasks/util"
	"time"
)

var (
	lastBlockHeightForCommitteeQ uint
)

func getCommitteeGASBalances(transfers []*models.Transfer) map[string]*big.Float {
	if len(transfers) == 0 {
		return nil
	}

	addresses := map[string]bool{}

	for _, transfer := range transfers {
		addresses[transfer.From] = true
		addresses[transfer.To] = true
	}

	queryAddr := []string{}
	minBlockIndex := transfers[0].BlockIndex
	committee := getCommittee(minBlockIndex)

	for _, committeeAddr := range committee {
		if addresses[committeeAddr] {
			continue
		}

		queryAddr = append(queryAddr, committeeAddr)
	}

	addrGASBalances := map[string]*big.Float{}

	if len(addrGASBalances) > 0 {
		time.Sleep(1 * time.Second)
	}

	for _, addr := range queryAddr {
		balance, ok := util.QueryNEP5Balance(minBlockIndex, addr, models.GAS, 8)
		if !ok {
			continue
		}

		addrGASBalances[addr] = balance
	}

	return addrGASBalances
}

func getCommittee(minBlockHeight uint) []string {
	if minBlockHeight < lastBlockHeightForCommitteeQ {
		return nil
	}

	lastBlockHeightForCommitteeQ = minBlockHeight

	invokeResult := rpc.GetCommittee(minBlockHeight)
	if invokeResult == nil ||
		util.VMStateFault(invokeResult.State) ||
		len(invokeResult.Stack) == 0 {
		return nil
	}

	stackItem := invokeResult.Stack[0]
	if stackItem.Type != "Array" {
		return nil
	}

	committee := []string{}

	for _, item := range stackItem.Value.([]interface{}) {
		stackItem := item.(map[string]interface{})
		addr, ok := util.GetAddressFromPublicKeyBase64(stackItem["value"].(string))
		if ok {
			committee = append(committee, addr)
		}
	}

	return committee
}
