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
	BlockIndex    uint
	BlockTime     uint64
	TxID          string
	Trigger       string
	VMState       string
	GasConsumed   *big.Float
	Stack         []StackItem
	Notifications []Notification
}

// Notification db model.
type Notification struct {
	ID         uint
	BlockIndex uint
	BlockTime  uint64
	TxID       string
	Trigger    string
	VMState    string
	Contract   string
	EventName  string
	State      *State
}

// State represents notification state.
type State struct {
	Type  string
	Value []StackItem
}

// StackItem represents value of a notification state.
type StackItem struct {
	Type  string
	Value interface{}
}

// MarshalStack is the shortcut of json.Marshal(appLog.Stack).
func (appLog *ApplicationLog) MarshalStack() []byte {
	stack, err := json.Marshal(appLog.Stack)
	if err != nil {
		log.Panic(err)
	}

	return stack
}

// UnmarshalStack is the shortcut of json.Unmarshal(stack, &appLog.Stack).
func (appLog *ApplicationLog) UnmarshalStack(stack []byte) {
	err := json.Unmarshal(stack, &appLog.Stack)
	if err != nil {
		log.Panic(err)
	}
}

// MarshalState is the shortcut of json.Marshal(noti.State).
func (noti *Notification) MarshalState() []byte {
	state, err := json.Marshal(noti.State)
	if err != nil {
		log.Panic(err)
	}

	return state
}

// UnmarshalState is the shortcut of json.Unmarshal(state, &noti.State).
func (noti *Notification) UnmarshalState(state []byte) {
	err := json.Unmarshal(state, &noti.State)
	if err != nil {
		log.Panic(err)
	}
}

// GetSrc returns source of this notification: block or tx.
func (noti *Notification) GetSrc() string {
	if noti.Trigger == "Application" {
		return "tx"
	}

	return "block"
}

// ParseApplicationLog parses struct raw application log rpc query result to db model.
func ParseApplicationLog(blockIndex uint, blockTime uint64, appLogResult *rpc.ApplicationLogResult) *ApplicationLog {
	appLog := ApplicationLog{
		TxID:        appLogResult.TxID,
		BlockIndex:  blockIndex,
		BlockTime:   blockTime,
		Trigger:     appLogResult.Trigger,
		VMState:     appLogResult.VMState,
		GasConsumed: appLogResult.GasConsumed,
		// Stack: parsed below.
	}

	appLog.Stack = make([]StackItem, len(appLogResult.Stack))
	for _, stack := range appLogResult.Stack {
		appLog.Stack = append(appLog.Stack, StackItem{
			Type:  stack.Type,
			Value: stack.Value,
		})
	}

	for _, notiResult := range appLogResult.Notifications {
		noti := Notification{
			BlockIndex: blockIndex,
			BlockTime:  blockTime,
			TxID:       appLogResult.TxID,
			Trigger:    appLog.Trigger,
			VMState:    appLogResult.VMState,
			Contract:   notiResult.Contract,
			EventName:  notiResult.EventName,
			State: &State{
				Type: notiResult.State.Type,
				// Value: parsed below.
			},
		}

		for _, value := range notiResult.State.Value {
			noti.State.Value = append(noti.State.Value, StackItem{
				Type:  value.Type,
				Value: value.Value,
			})
		}

		appLog.Notifications = append(appLog.Notifications, noti)
	}

	return &appLog
}

// ParseStackItem convert rpc.StackItem to models.StackItem.
func ParseStackItem(rawStackItem *rpc.StackItem) StackItem {
	return StackItem{
		Type:  rawStackItem.Type,
		Value: rawStackItem.Value,
	}
}
