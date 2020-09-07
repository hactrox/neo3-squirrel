package db

import (
	"database/sql"
	"fmt"
	"neo3-squirrel/models"
	"neo3-squirrel/pkg/mysql"
	"neo3-squirrel/util/log"
	"strings"
)

// InsertApplicationLog inserts applicationlog into database.
func InsertApplicationLog(appLog *models.ApplicationLog) {
	mysql.Trans(func(sqlTx *sql.Tx) error {
		if err := insertAppLogBasic(sqlTx, appLog); err != nil {
			return err
		}
		if err := insertAppLogNotifications(sqlTx, appLog.Notifications); err != nil {
			return err
		}

		return nil
	})
}

func insertAppLogBasic(sqlTx *sql.Tx, appLog *models.ApplicationLog) error {
	columns := []string{
		"`txid`",
		"`trigger`",
		"`vmstate`",
		"`gasconsumed`",
		"`stack`",
		"`notifications`",
	}

	query := []string{
		fmt.Sprintf("INSERT INTO `applicationlog` (%s)", strings.Join(columns, ", ")),
		fmt.Sprintf("VALUES(%s)", strings.Repeat(",?", len(columns))[1:]),
	}

	args := []interface{}{
		appLog.TxID,
		appLog.Trigger,
		appLog.VMState,
		fmt.Sprintf("%.8f", appLog.GasConsumed),
		appLog.Stack,
		len(appLog.Notifications),
	}

	if len(columns) != len(args) {
		log.Panicf("len(columns)=%d not equal to len(args)=%d", len(columns), len(args))
	}

	_, err := sqlTx.Exec(mysql.Compose(query), args...)
	if err != nil {
		log.Error(err)
	}

	return err
}

func insertAppLogNotifications(sqlTx *sql.Tx, notifications []models.Notification) error {
	columns := []string{
		`contract`,
		`eventname`,
		`state`,
	}

	var strBuilder strings.Builder
	strBuilder.WriteString(fmt.Sprintf("INSERT INTO `applicationlog_notification`"))
	strBuilder.WriteString(fmt.Sprintf("(%s)", strings.Join(columns, ", ")))
	strBuilder.WriteString("VALUES")

	// Construct (?, ?, ?) list.
	statement := fmt.Sprintf(",(%s)", strings.Repeat(",?", len(columns))[1:])
	strBuilder.WriteString(strings.Repeat(statement, len(notifications))[1:])

	// Construct sql query args.
	args := []interface{}{}
	for _, noti := range notifications {
		args = append(args, noti.Contract, noti.EventName, noti.State)
	}

	query := strBuilder.String()
	_, err := sqlTx.Exec(query, args...)
	if err != nil {
		log.Error(err)
	}

	return err
}
