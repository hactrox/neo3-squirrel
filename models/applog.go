package models

import (
	"encoding/json"
	"math/big"
	"neo3-squirrel/rpc"
	"neo3-squirrel/util/log"
)

// ApplicationLog db model.
type ApplicationLog struct {
	ID            uint
	TxID          string
	Trigger       string
	VMState       string
	GasConsumed   *big.Float
	Stack         []byte
	Notifications []Notification
}

// Notification db model.
type Notification struct {
	ID         uint
	TxID       string
	BlockIndex uint
	BlockTime  uint64
	VMState    string
	Contract   string
	EventName  string
	State      State
}

// State represents notification state.
type State struct {
	Type  string
	Value []StateValue
}

// StateValue represents value of a notification state.
type StateValue struct {
	Type  string
	Value interface{}
}

// ParseApplicationLog parses struct raw application log rpc query result to db model.
func ParseApplicationLog(tx *Transaction, appLogResult *rpc.ApplicationLogResult) ApplicationLog {
	stack, err := json.Marshal(appLogResult.Stack)
	if err != nil {
		log.Panic(err)
	}

	appLog := ApplicationLog{
		TxID:        appLogResult.TxID,
		Trigger:     appLogResult.Trigger,
		VMState:     appLogResult.VMState,
		GasConsumed: appLogResult.GasConsumed,
		Stack:       stack,
	}

	for _, notiResult := range appLogResult.Notifications {
		noti := Notification{
			TxID:       appLogResult.TxID,
			BlockIndex: tx.BlockIndex,
			BlockTime:  tx.BlockTime,
			VMState:    appLogResult.VMState,
			Contract:   notiResult.Contract,
			EventName:  notiResult.EventName,
			State: State{
				Type: notiResult.State.Type,
				// Value: notiResult.State.Value, being parsed below.
			},
		}

		for _, value := range notiResult.State.Value {
			noti.State.Value = append(noti.State.Value, StateValue{
				Type:  value.Type,
				Value: value.Value,
			})
		}

		appLog.Notifications = append(appLog.Notifications, noti)
	}

	return appLog
}
