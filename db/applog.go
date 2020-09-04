package db

import (
	"database/sql"
	"fmt"
	"neo3-squirrel/models"
	"neo3-squirrel/pkg/mysql"
	"neo3-squirrel/util/log"
	"strings"
)

// GetLastPKForApplicationLogTask returns the last
// primary key of the inserted application log record.
func GetLastPKForApplicationLogTask() uint {
	subQuery := []string{
		"SELECT `txid`",
		"FROM `applicationlog`",
		"ORDER BY `id` DESC",
		"LIMIT 1",
	}

	query := []string{
		"SELECT `id`",
		"FROM `transaction`",
		fmt.Sprintf("WHERE `hash` = (%s)", mysql.Compose(subQuery)),
		"LIMIT 1",
	}

	var pk uint
	err := mysql.QueryRow(mysql.Compose(query), nil, &pk)
	if err != nil {
		if mysql.IsRecordNotFoundError(err) {
			return 0
		}

		log.Error(mysql.Compose(query))
		log.Panic(err)
	}

	return pk
}

// InsertApplicationLog inserts applicationlog into database.
func InsertApplicationLog(appLog *models.ApplicationLog) {
	mysql.Trans(func(sqlTx *sql.Tx) error {
		insertAppLogBasic(sqlTx, appLog)
		insertAppLogNotifications(sqlTx, appLog.Notifications)
		return nil
	})
}

func insertAppLogBasic(sqlTx *sql.Tx, appLog *models.ApplicationLog) {
	columns := []string{
		"`txid`",
		"`trigger`",
		"`vmstate`",
		"`gasconsumed`",
		"`stack`",
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
	}

	if len(columns) != len(args) {
		log.Panicf("len(columns)=%d not equal to len(args)=%d", len(columns), len(args))
	}

	_, err := sqlTx.Exec(mysql.Compose(query), args...)
	if err != nil {
		log.Error(mysql.Compose(query))
		log.Panic(err)
	}
}

func insertAppLogNotifications(sqlTx *sql.Tx, notifications []models.Notification) {
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
		log.Error(query)
		log.Panic(err)
	}
}