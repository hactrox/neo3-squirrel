package util

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"neo3-squirrel/models"
	"neo3-squirrel/rpc"
	"neo3-squirrel/util/base58"
	"neo3-squirrel/util/convert"
	"neo3-squirrel/util/log"
	"strings"
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

	stackItem := models.StackItem{
		Type:  result.Stack[0].Type,
		Value: result.Stack[0].Value,
	}

	rawBalance, ok := extractValue(stackItem)
	if !ok {
		return nil, false
	}

	return convert.AmountReadable(rawBalance, decimals), true
}

// QueryNEP5Balances queries addresses balances from fullnode.
func QueryNEP5Balances(minBlockIndex uint, addresses []string, contract string, decimals uint) ([]*big.Float, bool) {
	if len(addresses) == 0 {
		err := fmt.Errorf("addresses cannot be empty")
		log.Panic(err)
	}

	script := ""
	for _, addr := range addresses {
		sc, err := generateNEP5BalanceOfScript(addr, contract)
		if err != nil {
			return nil, false
		}

		script += sc
	}

	// Encode script hex string to base64 encoding.
	scriptBytes, err := hex.DecodeString(script)
	if err != nil {
		log.Panic(err)
	}

	result := rpc.InvokeScript(minBlockIndex, base64.StdEncoding.EncodeToString(scriptBytes))
	if result == nil ||
		VMStateFault(result.State) ||
		len(result.Stack) < len(addresses) {
		return nil, false
	}

	readableBalances := []*big.Float{}

	for _, rawStackItem := range result.Stack {
		stackItem := models.StackItem{
			Type:  rawStackItem.Type,
			Value: rawStackItem.Value,
		}

		rawBalance, ok := extractValue(stackItem)
		if !ok {
			return nil, false
		}

		readableBalances = append(readableBalances, convert.AmountReadable(rawBalance, decimals))
	}

	return readableBalances, true
}

func generateNEP5BalanceOfScript(address, contract string) (string, error) {
	var strBuilder strings.Builder

	addrBytes, err := base58.CheckDecode(address)
	if err != nil {
		panic(err)
	}

	if len(addrBytes) < 21 {
		err := fmt.Errorf("invalid address: %s", address)
		log.Error(err)
		return "", err
	}

	addrBytes = addrBytes[1:21]

	contract = strings.TrimPrefix(contract, "0x")
	contractBytes, err := hex.DecodeString(contract)
	if err != nil {
		log.Error(contract)
		log.Error(err)
		return "", err
	}

	// 0c 14 [addrSC]
	// 11 c0
	// 0c 09 62616c616e63654f66(balanceOf)
	// 0c 14 [assetID]
	// 41 627d5b52(System.Contract.Call)

	strBuilder.WriteString("0c14")
	strBuilder.WriteString(hex.EncodeToString(addrBytes))
	strBuilder.WriteString("11c0")
	strBuilder.WriteString("0c0962616c616e63654f66")
	strBuilder.WriteString("0c14")
	strBuilder.WriteString(hex.EncodeToString(convert.ReverseBytes(contractBytes)))
	strBuilder.WriteString("41627d5b52")

	return strBuilder.String(), nil
}
