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

// InvokeFunction reflects the rpc call 'invokefunction'.
func InvokeFunction(minBlockIndex uint, contract, invokeFunc string, parameters []interface{}) *InvokeFunctionResult {
	params := []interface{}{contract, invokeFunc}
	if len(parameters) > 0 {
		params = append(params, parameters...)
	}

	const method = "invokefunction"
	args := generateRequestBody(method, params)
	resp := InvokefunctionResponse{}
	retryCnt := uint(0)
	delay := 0

	for {
		call(minBlockIndex, args, &resp)
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
