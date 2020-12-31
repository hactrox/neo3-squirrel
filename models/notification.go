package models

import (
	"encoding/json"
	"math/big"
	"neo3-squirrel/rpc"
	"neo3-squirrel/util/log"
)

// Trigger defiles application log trigger type.
type Trigger string

// Application sources.
const (
	SrcBlock = "block"
	SrcTx    = "tx"
)

// Triggers.
const (
	AppLogTriggerOnPersist    Trigger = "OnPersist"
	AppLogTriggerPostPersist  Trigger = "PostPersist"
	AppLogTriggerVerification Trigger = "Verification"
	AppLogTriggerApplication  Trigger = "Application"
	AppLogTriggerSystem       Trigger = "System"
	AppLogTriggerAll          Trigger = "All"
)

// Notification db model.
type Notification struct {
	ID          uint
	BlockIndex  uint
	BlockTime   uint64
	Hash        string
	Src         string
	ExecIndex   uint
	Trigger     string
	VMState     string
	Exception   string
	GasConsumed *big.Float
	Stack       []StackItem
	N           uint
	Contract    string
	EventName   string
	State       *State
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

// MarshalStack is the shortcut of json.Marshal(noti.Stack).
func (noti *Notification) MarshalStack() []byte {
	stack, err := json.Marshal(noti.Stack)
	if err != nil {
		log.Panic(err)
	}

	return stack
}

// UnmarshalStack is the shortcut of json.Unmarshal(stack, &noti.Stack).
func (noti *Notification) UnmarshalStack(stack []byte) {
	err := json.Unmarshal(stack, &noti.Stack)
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

// ParseApplicationLog parses struct raw application log rpc query result into []*models.Notification.
func ParseApplicationLog(blockIndex uint, blockTime uint64, appLog *rpc.ApplicationLog) []*Notification {
	hash, src := getHashSrc(appLog)

	notis := []*Notification{}

	for execIdx, exec := range appLog.Executions {
		for notiIdx, rawNoti := range exec.Notifications {
			noti := Notification{
				BlockIndex:  blockIndex,
				BlockTime:   blockTime,
				Hash:        hash,
				Src:         src,
				ExecIndex:   uint(execIdx),
				Trigger:     exec.Trigger,
				VMState:     exec.VMState,
				Exception:   exec.Exception,
				GasConsumed: exec.GasConsumed,
				Stack:       parseNotiStack(exec),
				N:           uint(notiIdx),
				Contract:    rawNoti.Contract,
				EventName:   rawNoti.EventName,
				State:       parseNotiState(rawNoti),
			}

			notis = append(notis, &noti)
		}
	}

	return notis
}

func getHashSrc(appLog *rpc.ApplicationLog) (string, string) {
	if appLog.BlockHash != "" {
		return appLog.BlockHash, SrcBlock
	}

	return appLog.TxID, SrcTx
}

func parseNotiStack(exec rpc.AppLogExecution) []StackItem {
	stackArr := make([]StackItem, len(exec.Stack))

	for i, stack := range exec.Stack {
		stackArr[i] = ParseStackItem(&stack)
	}

	return stackArr
}

func parseNotiState(rawNoti rpc.Notification) *State {
	state := &State{
		Type: rawNoti.State.Type,
		// Value: parsed below.
	}

	for _, value := range rawNoti.State.Value {
		state.Value = append(state.Value, ParseStackItem(&value))
	}

	return state
}

// ParseStackItem convert rpc.StackItem to models.StackItem.
func ParseStackItem(rawStackItem *rpc.StackItem) StackItem {
	return StackItem{
		Type:  rawStackItem.Type,
		Value: rawStackItem.Value,
	}
}
