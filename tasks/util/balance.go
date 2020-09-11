package util

import (
	"fmt"
	"math/big"
	"neo3-squirrel/models"
	"neo3-squirrel/rpc"
	"neo3-squirrel/util/convert"
	"neo3-squirrel/util/log"
)

// QueryNEP5Balance queries address contract balance from fullnode.
func QueryNEP5Balance(minBlockIndex uint, address, contract string, decimals uint) (*big.Float, bool) {
	if len(address) == 0 {
		err := fmt.Errorf("address cannot be empty")
		log.Panic(err)
	}

	const method = "balanceOf"
	addrScriptHash := GetAddrScriptHash(address)
	params := []interface{}{
		[]rpc.StackItem{
			{Type: "Hash160", Value: addrScriptHash},
		},
	}

	result := rpc.InvokeFunction(minBlockIndex, contract, method, params)
	if result == nil ||
		VMStateFault(result.State) ||
		len(result.Stack) == 0 {
		return nil, false
	}

	// log.Debug(result)
	stackItem := models.StackItem{
		Type:  result.Stack[0].Type,
		Value: result.Stack[0].Value,
	}

	balance, ok := extractValue(stackItem)
	if !ok {
		return nil, false
	}

	return convert.AmountReadable(balance, decimals), true
}
