package rpc

import (
	"math/big"
	"math/rand"
	"neo3-squirrel/util/log"
	"time"
)

// InvokefunctionResponse is the response structure of rpc call 'invokefunction'.
type InvokefunctionResponse struct {
	responseCommon
	Result *InvokeFunctionResult `json:"result"`
}

// ContractStatesResponse is the response structure of rpc call 'getcontractstates'.
type ContractStatesResponse struct {
	responseCommon
	Result []*ContractState `json:"result"`
}

// InvokeFunctionResult represents invokefunction query result.
type InvokeFunctionResult struct {
	Script      string      `json:"script"`
	State       string      `json:"state"`
	GasConsumed *big.Float  `json:"gasconsumed"`
	Stack       []StackItem `json:"stack"`
	Tx          string      `json:"tx"`
}

// StackItem represents value of a notification state.
type StackItem struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// ContractState represents 'getcontractstates' query result.
type ContractState struct {
	ID          int         `json:"id"`
	Hash        string      `json:"hash"`
	Script      string      `json:"script"`
	Manifest    interface{} `json:"manifest"`
	BlockIndex  uint        `json:"blockindex"`
	BlockTime   uint64      `json:"blocktime"`
	State       string      `json:"state"`
	TxID        string      `json:"txid"`
	Name        string      `json:"name"`
	Symbol      string      `json:"symbol"`
	Decimals    uint        `json:"decimals"`
	TotalSupply *big.Float  `json:"totalSupply"`
}

// InvokeFunction reflects the rpc call 'invokefunction'.
func InvokeFunction(minBlockIndex uint, contract, invokeFunc string, parameters []interface{}) *InvokeFunctionResult {
	const method = "invokefunction"
	params := []interface{}{contract, invokeFunc}
	if len(parameters) > 0 {
		params = append(params, parameters...)
	}

	args := generateRequestBody(method, params)
	resp := InvokefunctionResponse{}
	retryCnt := uint(0)
	delay := 0

	for {
		request(minBlockIndex, args, &resp)
		if resp.Result != nil {
			return resp.Result
		}

		if resp.Error != nil {
			log.Warnf("Invalid '%s' call: params=%s", method, args)
		}

		retryCnt++
		if delay < 10*1000 {
			delay = rand.Intn(1<<retryCnt) + 1000
		}

		log.Warnf("Failed to invoke smartcontract func %s of contract %s. Delay for %d msecs and retry(retry=%d).", invokeFunc, contract, delay, retryCnt)

		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
}

// GetContractStates reflects the rpc call 'getcontractstates'.
func GetContractStates(fromBlockIndex, batches uint) []*ContractState {
	const method = "getcontractstates"
	params := []interface{}{fromBlockIndex, batches}

	args := generateRequestBody(method, params)
	resp := ContractStatesResponse{}
	request(fromBlockIndex, args, &resp)
	return resp.Result
}
