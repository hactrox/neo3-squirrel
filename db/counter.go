package db

import (
	"database/sql"
	"fmt"
	"neo3-squirrel/pkg/mysql"
	"neo3-squirrel/util/log"
	"os"
)

// Counter db model.
type Counter struct {
	ID         uint
	BlockIndex int

	AddrCount uint
}

// GetLastBlockHeight returns the last block height persisted.
func GetLastBlockHeight() int {
	return getCounterInstance().BlockIndex
}

func getCounterInstance() Counter {
	query := []string{
		"SELECT `id`, `block_index`, `addr_count`",
		"FROM `counter`",
		"WHERE `id` = 1",
		"LIMIT 1",
	}

	var counter Counter

	err := mysql.QueryRow(mysql.Compose(query), nil,
		&counter.ID,
		&counter.BlockIndex,
		&counter.AddrCount,
	)

	if err != nil {
		if mysql.IsRecordNotFoundError(err) {
			log.Error("DB table 'counter' not initialized.")
			os.Exit(1)
		}

		log.Panic(err)
	}

	return counter
}

func updateCounter(sqlTx *sql.Tx, field string, value interface{}) error {
	query := []string{
		"UPDATE `counter`",
		fmt.Sprintf("SET %s=%v", field, value),
		"WHERE `id`=1",
		"LIMIT 1",
	}

	result, err := sqlTx.Exec(mysql.Compose(query))
	if err != nil {
		log.Error(err)
		return err
	}

	mysql.CheckIfRowsNotAffected(result, query)
	return err
}
