package util

import (
	"log"
	"math/big"
	"neo3-squirrel/models"
	"neo3-squirrel/rpc"
	"strings"
)

// QueryAssetBasicInfo gets contract name, symbol, decimals and totalSupply from fullnode.
func QueryAssetBasicInfo(minBlockIndex uint, asset *models.Asset) bool {
	contract := asset.Contract
	var ok bool

	if len(contract) == 0 {
		log.Panic("asset contract cannot be empty before query")
	}

	asset.Name, ok = queryContractName(minBlockIndex, contract)
	if !ok {
		return false
	}
	asset.Symbol, ok = queryContractSymbol(minBlockIndex, contract)
	if !ok {
		return false
	}

	asset.Decimals, ok = queryContractDecimals(minBlockIndex, contract)
	if !ok {
		return false
	}

	asset.TotalSupply, ok = queryContractTotalSupply(minBlockIndex, contract)
	if !ok {
		return false
	}

	return true
}

func queryContractName(minBlockIndex uint, contract string) (string, bool) {
	stack, ok := queryContractProperty(minBlockIndex, contract, "name")
	if !ok {
		return "", false
	}

	return extractString(stack.Type, stack.Value)
}

func queryContractSymbol(minBlockIndex uint, contract string) (string, bool) {
	stack, ok := queryContractProperty(minBlockIndex, contract, "symbol")
	if !ok {
		return "", false
	}

	return extractString(stack.Type, stack.Value)
}

func queryContractDecimals(minBlockIndex uint, contract string) (*big.Float, bool) {
	stack, ok := queryContractProperty(minBlockIndex, contract, "decimals")
	if !ok {
		return nil, false
	}

	return extractValue(stack.Type, stack.Value)
}

func queryContractTotalSupply(minBlockIndex uint, contract string) (*big.Float, bool) {
	stack, ok := queryContractProperty(minBlockIndex, contract, "totalSupply")
	if !ok {
		return nil, false
	}

	return extractValue(stack.Type, stack.Value)
}

// name, symbol...
func queryContractProperty(minBlockIndex uint, contract, property string) (*rpc.StackItem, bool) {
	result := rpc.InvokeFunction(minBlockIndex, contract, property, nil)
	if result == nil ||
		strings.Contains(result.State, "FAULT") ||
		len(result.Stack) == 0 {
		return nil, false
	}

	return &result.Stack[0], true
}
