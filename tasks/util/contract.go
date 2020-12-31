package util

import (
	"encoding/base64"
	"encoding/hex"
	"math/big"
	"neo3-squirrel/models"
	"neo3-squirrel/rpc"
	"neo3-squirrel/util/byteutil"
	"neo3-squirrel/util/convert"
	"neo3-squirrel/util/log"
	"strings"
)

// GetContractHash extracts target contract hash from management contract noti state value.
func GetContractHash(csNoti *models.Notification) (string, bool) {
	state := csNoti.State
	if state.Type != "Array" ||
		len(state.Value) == 0 {
		return "", false
	}

	contractHashBase64, ok := state.Value[0].Value.(string)
	if !ok {
		return "", false
	}

	contractHash, err := base64.StdEncoding.DecodeString(contractHashBase64)
	if err != nil {
		log.Panic(err)
	}

	return "0x" + hex.EncodeToString(byteutil.ReverseBytes(contractHash)), true
}

// QueryAssetBasicInfo gets contract symbol, decimals and totalSupply from fullnode.
func QueryAssetBasicInfo(minBlockIndex uint, asset *models.Asset) bool {
	contract := asset.Contract
	var ok bool

	if len(contract) == 0 {
		log.Panic("asset contract cannot be empty before query")
	}

	asset.Symbol, ok = queryContractSymbol(minBlockIndex, contract)
	if !ok {
		log.Warnf("Failed to get 'symbol' from contract %s", contract)
		return false
	}

	decimals, ok := queryContractDecimals(minBlockIndex, contract)
	if !ok {
		log.Warnf("Failed to get 'decimals' from contract %s", contract)
		return false
	}

	dec, accuracy := decimals.Int64()
	if accuracy != big.Exact {
		return false
	}
	asset.Decimals = uint(dec)

	asset.TotalSupply, ok = QueryAssetTotalSupply(minBlockIndex, contract, asset.Decimals)
	if !ok {
		log.Warnf("Failed to get 'totalSupply' from contract %s", contract)
	}

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

// query symbol and decimals
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
