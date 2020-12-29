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

var appLogNotiColumns = []string{
	"`id`",
	"`block_index`",
	"`block_time`",
	"`hash`",
	"`src`",
	"`exec_idx`",
	"`trigger`",
	"`vmstate`",
	"`gasconsumed`",
	"`stack`",
	"`n`",
	"`contract`",
	"`eventname`",
	"`state`",
}

// InsertAppLogNotifications inserts applicationlog notifications into database.
func InsertAppLogNotifications(notifications []*models.Notification) {
	mysql.Trans(func(sqlTx *sql.Tx) error {
		if err := insertAppLogNotifications(sqlTx, notifications); err != nil {
			return err
		}

		return nil
	})
}

func insertAppLogNotifications(sqlTx *sql.Tx, notifications []*models.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	var strBuilder strings.Builder
	strBuilder.WriteString("INSERT INTO `notification`")
	strBuilder.WriteString(fmt.Sprintf("(%s)", strings.Join(appLogNotiColumns[1:], ", ")))
	strBuilder.WriteString("VALUES")

	// Construct (?, ?, ?) list.
	statement := fmt.Sprintf(",(%s)", strings.Repeat(",?", len(appLogNotiColumns[1:]))[1:])
	strBuilder.WriteString(strings.Repeat(statement, len(notifications))[1:])

	// Construct sql query args.
	args := []interface{}{}
	for _, noti := range notifications {
		args = append(args,
			noti.BlockIndex,
			noti.BlockTime,
			noti.Hash,
			noti.Src,
			noti.ExecIndex,
			noti.Trigger,
			noti.VMState,
			convert.BigFloatToString(noti.GasConsumed),
			noti.MarshalStack(),
			noti.N,
			noti.Contract,
			noti.EventName,
			noti.MarshalState(),
		)
	}

	query := strBuilder.String()
	_, err := sqlTx.Exec(query, args...)
	if err != nil {
		log.Error(err)
	}

	return err
}

// GetLastNotification returns the last notification record.
func GetLastNotification() *models.Notification {
	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(appLogNotiColumns, ", ")),
		"FROM `notification`",
		"ORDER BY `id` DESC",
		"LIMIT 1",
	}

	return getNotiQueryRow(query)
}

// GetLastNotiForNEP5Task returns the last notification
// of the NEP5 transfer record.
func GetLastNotiForNEP5Task() *models.Notification {
	subQuery := []string{
		"SELECT `hash`",
		"FROM `transfer`",
		"ORDER BY `id` DESC",
		"LIMIT 1",
	}

	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(appLogNotiColumns, ", ")),
		"FROM `notification`",
		fmt.Sprintf("WHERE `hash` = (%s)", mysql.Compose(subQuery)),
		"ORDER BY `id` DESC",
		"LIMIT 1",
	}

	return getNotiQueryRow(query)
}

// GetNotificationsGroupedByTxID returns notifications grouped by txid.
func GetNotificationsGroupedByTxID(startPK, groups uint) []*models.Notification {
	// SELECT * FROM `notification`
	// WHERE `id` >= {startPK} AND `hash` IN (
	// 	SELECT `hash` FROM (
	// 		SELECT DISTINCT `hash` FROM `notification`
	// 		WHERE `id` >= {startPK}
	// 		ORDER BY `id`
	// 		LIMIT {limit}
	// 	) a
	// );

	subQuery := []string{
		"SELECT `hash`",
		"FROM (",
		"SELECT DISTINCT `hash`",
		"FROM `notification`",
		fmt.Sprintf("WHERE `id` >= %d", startPK),
		fmt.Sprintf("LIMIT %d", groups),
		") a",
	}

	query := []string{
		fmt.Sprintf("SELECT %s", strings.Join(appLogNotiColumns, ", ")),
		"FROM `notification`",
		fmt.Sprintf("WHERE `id` >= %d", startPK),
		fmt.Sprintf("AND `hash` IN (%s)", mysql.Compose(subQuery)),
	}

	return getAppLogNotiQuery(query)
}

// GetNotificationCount returns the number of notifications
// starts from the given primary key(>=startPK);
func GetNotificationCount(startPK uint) uint {
	query := []string{
		"SELECT COUNT(`id`)",
		"FROM `notification`",
		fmt.Sprintf("WHERE `id` >= %d", startPK),
	}

	var count uint
	err := mysql.QueryRow(mysql.Compose(query), nil, &count)
	if err != nil {
		log.Error(mysql.Compose(query))
		log.Panic(err)
	}

	return count
}

/* ------------------------------
	DB query result parser
------------------------------ */

func getNotiQueryRow(query []string) *models.Notification {
	var noti models.Notification
	gasConsumedStr := ""
	stack := []byte{}
	state := []byte{}

	err := mysql.QueryRow(mysql.Compose(query), nil,
		&noti.ID,
		&noti.BlockIndex,
		&noti.BlockTime,
		&noti.Hash,
		&noti.Src,
		&noti.ExecIndex,
		&noti.Trigger,
		&noti.VMState,
		&gasConsumedStr,
		&stack,
		&noti.N,
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

	noti.UnmarshalStack(stack)
	noti.UnmarshalState(state)

	return &noti
}

func getAppLogNotiQuery(query []string) []*models.Notification {
	rows, err := mysql.Query(mysql.Compose(query))
	if err != nil {
		log.Error(mysql.Compose(query))
		log.Panic(err)
	}

	defer rows.Close()
	notifications := []*models.Notification{}

	for rows.Next() {
		var noti models.Notification
		gasConsumedStr := ""
		stack := []byte{}
		state := []byte{}

		err := rows.Scan(
			&noti.ID,
			&noti.BlockIndex,
			&noti.BlockTime,
			&noti.Hash,
			&noti.Src,
			&noti.ExecIndex,
			&noti.Trigger,
			&noti.VMState,
			&gasConsumedStr,
			&stack,
			&noti.N,
			&noti.Contract,
			&noti.EventName,
			&state,
		)
		if err != nil {
			log.Panic(err)
		}

		noti.UnmarshalStack(stack)
		noti.UnmarshalState(state)

		notifications = append(notifications, &noti)
	}

	return notifications
}
