package util

import (
	"math/big"
	"neo3-squirrel/models"
	"neo3-squirrel/rpc"
	"neo3-squirrel/util/convert"
	"neo3-squirrel/util/log"
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

	decimals, ok := queryContractDecimals(minBlockIndex, contract)
	if !ok {
		return false
	}

	dec, accuracy := decimals.Int64()
	if accuracy != big.Exact {
		return false
	}
	asset.Decimals = uint(dec)

	asset.TotalSupply, ok = QueryAssetTotalSupply(minBlockIndex, contract, asset.Decimals)
	return ok
}

// QueryAssetTotalSupply queries total supply of the given contract and returns as decimals-formatted value.
func QueryAssetTotalSupply(minBlockIndex uint, contract string, decimals uint) (*big.Float, bool) {
	totalSupply, ok := queryContractTotalSupply(minBlockIndex, contract)
	if !ok {
		return nil, false
	}

	totalSupply = convert.AmountReadable(totalSupply, decimals)
	return totalSupply, true
}

func queryContractName(minBlockIndex uint, contract string) (string, bool) {
	stack, ok := queryContractProperty(minBlockIndex, contract, "name")
	if !ok {
		return "", false
	}

	return extractString(models.ParseStackItem(stack))
}

func queryContractSymbol(minBlockIndex uint, contract string) (string, bool) {
	stack, ok := queryContractProperty(minBlockIndex, contract, "symbol")
	if !ok {
		return "", false
	}

	return extractString(models.ParseStackItem(stack))
}

func queryContractDecimals(minBlockIndex uint, contract string) (*big.Float, bool) {
	stack, ok := queryContractProperty(minBlockIndex, contract, "decimals")
	if !ok {
		return nil, false
	}

	return extractValue(models.ParseStackItem(stack))
}

func queryContractTotalSupply(minBlockIndex uint, contract string) (*big.Float, bool) {
	stack, ok := queryContractProperty(minBlockIndex, contract, "totalSupply")
	if !ok {
		return nil, false
	}

	return extractValue(models.ParseStackItem(stack))
}

// name, symbol...
func queryContractProperty(minBlockIndex uint, contract, property string) (*rpc.StackItem, bool) {
	result := rpc.InvokeFunction(minBlockIndex, contract, property, nil)
	if result == nil ||
		VMStateFault(result.State) ||
		len(result.Stack) == 0 {
		return nil, false
	}

	return &result.Stack[0], true
}

// VMStateFault tells if vm state fault.
func VMStateFault(vmstate string) bool {
	return strings.Contains(vmstate, "FAULT")
}
