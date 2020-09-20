package db

import (
	"database/sql"
	"fmt"
	"neo3-squirrel/models"
	"neo3-squirrel/pkg/mysql"
	"neo3-squirrel/util/convert"
	"neo3-squirrel/util/log"
	"strings"
)

var appLogColumns = []string{
	"`id`",
	"`block_index`",
	"`block_time`",
	"`txid`",
	"`trigger`",
	"`vmstate`",
	"`gasconsumed`",
	"`stack`",
	"`notifications`",
}

var appLogNotiColumns = []string{
	"`id`",
	"`block_index`",
	"`block_time`",
	"`txid`",
	"`vmstate`",
	"`contract`",
	"`eventname`",
	"`state`",
}

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

// GetApplicationLogByID returns application log by primary key.
func GetApplicationLogByID(pk uint) *models.ApplicationLog {
	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(appLogColumns, ", ")),
		"FROM `applicationlog`",
		fmt.Sprintf("WHERE `id` = %d", pk),
		"LIMIT 1",
	}

	return getApplicatoinLogQuery(mysql.Compose(query))
}

// GetApplicationLogByTxID returns application log by txID.
func GetApplicationLogByTxID(txID string) *models.ApplicationLog {
	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(appLogColumns, ", ")),
		"FROM `applicationlog`",
		fmt.Sprintf("WHERE `txid` = '%s'", txID),
		"LIMIT 1",
	}

	return getApplicatoinLogQuery(mysql.Compose(query))
}

func getApplicatoinLogQuery(query string) *models.ApplicationLog {
	appLog := models.ApplicationLog{}
	var gasConsumed string
	var notifications uint
	var stack []byte

	err := mysql.QueryRow(query, nil,
		&appLog.ID,
		&appLog.BlockIndex,
		&appLog.BlockTime,
		&appLog.TxID,
		&appLog.Trigger,
		&appLog.VMState,
		&gasConsumed,
		&stack,
		&notifications,
	)
	if err != nil {
		if mysql.IsRecordNotFoundError(err) {
			return nil
		}

		log.Error(query)
		log.Panic(err)
	}

	appLog.GasConsumed = convert.ToDecimal(gasConsumed)
	appLog.UnmarshalStack(stack)

	return &appLog
}

// GetLastNotiForNEP5Task returns the last notification
// of the NEP5 transfer record.
func GetLastNotiForNEP5Task() *models.Notification {
	subQuery := []string{
		"SELECT `txid`",
		"FROM `transfer`",
		"ORDER BY `id` DESC",
		"LIMIT 1",
	}

	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(appLogNotiColumns, ", ")),
		"FROM `applicationlog_notification`",
		fmt.Sprintf("WHERE `txid` = (%s)", mysql.Compose(subQuery)),
		"LIMIT 1",
	}

	var noti models.Notification
	state := []byte{}

	err := mysql.QueryRow(mysql.Compose(query), nil,
		&noti.ID,
		&noti.BlockIndex,
		&noti.BlockTime,
		&noti.TxID,
		&noti.VMState,
		&noti.Contract,
		&noti.EventName,
		&state,
	)
	if err != nil {
		if mysql.IsRecordNotFoundError(err) {
			return nil
		}

		log.Error(mysql.Compose(query))
		log.Panic(err)
	}

	noti.UnmarshalState(state)

	return &noti
}

// GetGroupedAppLogNotifications returns grouped notifications
// of a set of application logs.
// NOTE: ALL VMSTATE(NONE, HALT, FAULT, BREAK) INCLUDED.
func GetGroupedAppLogNotifications(appLogPK, limit uint) []*models.Notification {
	// SELECT *
	// FROM `applicationlog_notification`
	// WHERE `txid` IN (
	// 	SELECT `txid` FROM (
	// 		SELECT `txid` FROM `applicationlog`
	// 		WHERE `id` > {appLogPK} `limit` {limit}
	// 	) `tbl`
	// );

	subQuery := []string{
		"SELECT `txid`",
		"FROM (SELECT `txid`",
		"FROM `applicationlog`",
		fmt.Sprintf("WHERE `id` >= %d", appLogPK),
		fmt.Sprintf("LIMIT %d) `tbl`", limit),
	}

	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(appLogNotiColumns, ", ")),
		"FROM `applicationlog_notification`",
		fmt.Sprintf("WHERE `txid` IN (%s)", mysql.Compose(subQuery)),
	}

	rows, err := mysql.Query(mysql.Compose(query))
	if err != nil {
		log.Error(mysql.Compose(query))
		log.Panic(err)
	}

	defer rows.Close()
	notifications := []*models.Notification{}

	for rows.Next() {
		var noti models.Notification
		state := []byte{}

		err := rows.Scan(
			&noti.ID,
			&noti.BlockIndex,
			&noti.BlockTime,
			&noti.TxID,
			&noti.VMState,
			&noti.Contract,
			&noti.EventName,
			&state,
		)
		if err != nil {
			log.Panic(err)
		}

		noti.UnmarshalState(state)

		notifications = append(notifications, &noti)
	}

	return notifications
}

// GetAppLogNotifications gets application log notifications starts from
// the given primary key(>={startPK}) and limits to {limit} record from database.
func GetAppLogNotifications(startPK, limit uint) []*models.Notification {
	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(appLogNotiColumns, ", ")),
		"FROM `applicationlog_notification`",
		fmt.Sprintf("WHERE `id` > %d", startPK),
		fmt.Sprintf("LIMIT %d", limit),
	}

	rows, err := mysql.Query(mysql.Compose(query))
	if err != nil {
		log.Error(mysql.Compose(query))
		log.Panic(err)
	}

	defer rows.Close()
	notifications := []*models.Notification{}

	for rows.Next() {
		var noti models.Notification
		state := []byte{}

		err := rows.Scan(
			&noti.ID,
			&noti.BlockIndex,
			&noti.BlockTime,
			&noti.TxID,
			&noti.VMState,
			&noti.Contract,
			&noti.EventName,
			&state,
		)
		if err != nil {
			log.Panic(err)
		}

		noti.UnmarshalState(state)

		notifications = append(notifications, &noti)
	}

	return notifications
}

// GetNotificationCount returns the number of notifications
// starts from the given primary key(>=startPK);
func GetNotificationCount(startPK uint, txID string) uint {
	query := []string{
		"SELECT COUNT(`id`)",
		"FROM `applicationlog_notification`",
		fmt.Sprintf("WHERE `id` >= %d AND `txid` != '%s'", startPK, txID),
	}

	var count uint
	err := mysql.QueryRow(mysql.Compose(query), nil, &count)
	if err != nil {
		log.Error(mysql.Compose(query))
		log.Panic(err)
	}

	return count
}

func insertAppLogBasic(sqlTx *sql.Tx, appLog *models.ApplicationLog) error {
	columns := appLogColumns[1:]
	query := []string{
		fmt.Sprintf("INSERT INTO `applicationlog` (%s)", strings.Join(columns, ", ")),
		fmt.Sprintf("VALUES(%s)", strings.Repeat(",?", len(columns))[1:]),
	}

	args := []interface{}{
		appLog.BlockIndex,
		appLog.BlockTime,
		appLog.TxID,
		appLog.Trigger,
		appLog.VMState,
		convert.BigFloatToString(appLog.GasConsumed),
		appLog.MarshalStack(),
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
	if len(notifications) == 0 {
		return nil
	}

	var strBuilder strings.Builder
	strBuilder.WriteString("INSERT INTO `applicationlog_notification`")
	strBuilder.WriteString(fmt.Sprintf("(%s)", strings.Join(appLogNotiColumns[1:], ", ")))
	strBuilder.WriteString("VALUES")

	// Construct (?, ?, ?) list.
	statement := fmt.Sprintf(",(%s)", strings.Repeat(",?", len(appLogNotiColumns[1:]))[1:])
	strBuilder.WriteString(strings.Repeat(statement, len(notifications))[1:])

	// Construct sql query args.
	args := []interface{}{}
	for _, noti := range notifications {
		state := noti.MarshalState()

		args = append(args,
			noti.BlockIndex,
			noti.BlockTime,
			noti.TxID,
			noti.VMState,
			noti.Contract,
			noti.EventName,
			state,
		)
	}

	query := strBuilder.String()
	_, err := sqlTx.Exec(query, args...)
	if err != nil {
		log.Error(err)
	}

	return err
}
