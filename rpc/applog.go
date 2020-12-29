package rpc

import (
	"math/big"
	"math/rand"
	"neo3-squirrel/util/log"
	"time"
)

// ApplicationLogResponse is the response structure of rpc call 'getapplicationlog'.
type ApplicationLogResponse struct {
	responseCommon
	Result *ApplicationLogResult `json:"result"`
}

// ApplicationLogResult represents applicationlog query result.
type ApplicationLogResult struct {
	BlockHash  string            `json:"blockhash"`
	TxID       string            `json:"txid"`
	Executions []AppLogExecution `json:"executions"`
}

// AppLogExecution represents applicationlog executions.
type AppLogExecution struct {
	Trigger       string         `json:"trigger"`
	VMState       string         `json:"vmstate"`
	GasConsumed   *big.Float     `json:"gasconsumed"`
	Stack         []StackItem    `json:"stack"`
	Notifications []Notification `json:"notifications"`
}

// Notification represents a single notification of applicatino log.
type Notification struct {
	Contract  string `json:"contract"`
	EventName string `json:"eventname"`
	State     State  `json:"state"`
}

// State represents notification state.
type State struct {
	Type  string      `json:"type"`
	Value []StackItem `json:"value"`
}

// GetApplicationLog reflects the rpc call 'getapplicationlog'.
func GetApplicationLog(minBlockIndex uint, txID string) *ApplicationLogResult {
	params := []interface{}{txID}
	const method = "getapplicationlog"
	args := generateRequestBody(method, params)
	resp := ApplicationLogResponse{}
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

		log.Warnf("Cannot get applicationlog of txID: %s. Delay for %d msecs and retry(retry=%d).", txID, delay, retryCnt)

		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
}
