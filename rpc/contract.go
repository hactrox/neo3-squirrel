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

// ContractStatesResponse is the response structure of rpc call 'getcontractstates'.
type ContractStatesResponse struct {
	responseCommon
	Result *ContractState `json:"result"`
}

// ContractState represents 'getcontractstates' query result.
type ContractState struct {
	ID            int              `json:"id"`
	UpdateCounter uint             `json:"updatecounter"`
	Hash          string           `json:"hash"`
	NEF           ContractNEF      `json:"nef"`
	Manifest      ContractManifest `json:"manifest"`
}

type ContractNEF struct {
	Magic    uint64      `json:"magic"`
	Compiler string      `json:"compiler"`
	Tokens   interface{} `json:"tokens"`
	Script   string      `json:"script"`
	CheckSum uint64      `json:"checksum"`
}

// ContractManifest represents the manifest struct of contract state.
type ContractManifest struct {
	Name               string      `json:"name"`
	Groups             interface{} `json:"groups"`
	Features           interface{} `json:"features"`
	SupportedStandards []string    `json:"supportedstandards"`
	ABI                interface{} `json:"abi"`
	Permissions        interface{} `json:"permissions"`
	Trusts             interface{} `json:"trusts"`
	Extra              interface{} `json:"extra"`
}

// InvokeScript reflects the rpc call 'invokescript'.
func InvokeScript(minBlockIndex uint, script string) *InvokeFunctionResult {
	const method = "invokescript"
	params := []interface{}{script}

	return doInvoke(minBlockIndex, method, params)
}

// InvokeFunction reflects the rpc call 'invokefunction'.
func InvokeFunction(minBlockIndex uint, contract, invokeFunc string, parameters []interface{}) *InvokeFunctionResult {
	const method = "invokefunction"
	params := []interface{}{contract, invokeFunc}
	if len(parameters) > 0 {
		params = append(params, parameters...)
	}

	return doInvoke(minBlockIndex, method, params)
}

func doInvoke(minBlockIndex uint, method string, params []interface{}) *InvokeFunctionResult {
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

		log.Warnf("Failed to invoke smartcontract: %v. Delay for %d msecs and retry(retry=%d).", args, delay, retryCnt)

		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
}

// GetContractState reflects the rpc call 'getcontractstate'.
func GetContractState(fromBlockIndex uint, hash string) *ContractState {
	const method = "getcontractstate"
	params := []interface{}{hash}

	args := generateRequestBody(method, params)
	resp := ContractStatesResponse{}
	request(fromBlockIndex, args, &resp)
	return resp.Result
}
