package models

import (
	"encoding/json"
	"log"
	"math/big"
	"neo3-squirrel/rpc"
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
	Contract  string
	EventName string
	State     []byte
}

// ParseApplicationLog parses struct raw application log rpc query result to db model.
func ParseApplicationLog(appLogResult *rpc.ApplicationLogResult) ApplicationLog {
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
		state, err := json.Marshal(notiResult.State)
		if err != nil {
			log.Panic(err)
		}

		noti := Notification{
			Contract:  notiResult.Contract,
			EventName: notiResult.EventName,
			State:     state,
		}

		appLog.Notifications = append(appLog.Notifications, noti)
	}

	return appLog
}
