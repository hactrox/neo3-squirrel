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
	ID             uint
	BlockIndex     int
	ContractNotiPK uint

	AddrCount uint
}

/* ------------------------------
			Getter
------------------------------ */

// GetLastBlockHeight returns the last block height persisted.
func GetLastBlockHeight() int {
	return getCounterInstance().BlockIndex
}

// GetContractNotiPK returns the last contract notification primary key persisted.
func GetContractNotiPK() uint {
	return getCounterInstance().ContractNotiPK
}

// UpdateContractNotiPK updates `contract_noti_pk` counter.
func UpdateContractNotiPK(pk uint) error {
	return mysql.Trans(func(sqlTx *sql.Tx) error {
		return updateContractNotiPK(sqlTx, pk)
	})
}

func getCounterInstance() Counter {
	query := []string{
		"SELECT `id`, `block_index`, `contract_noti_pk`, `addr_count`",
		"FROM `counter`",
		"WHERE `id` = 1",
		"LIMIT 1",
	}

	var counter Counter

	err := mysql.QueryRow(mysql.Compose(query), nil,
		&counter.ID,
		&counter.BlockIndex,
		&counter.ContractNotiPK,
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

/* ------------------------------
			Updater
------------------------------ */

func updateBlockIndexCounter(sqlTx *sql.Tx, blockIndex uint) error {
	return updateCounter(sqlTx, "`block_index`", blockIndex)
}

func updateAddressCounter(sqlTx *sql.Tx, addrAdded uint) error {
	return updateCounterDelta(sqlTx, "`addr_count`", int64(addrAdded))
}

func updateTxCounter(sqlTx *sql.Tx, txAdded int) error {
	return updateCounterDelta(sqlTx, "`tx_count`", int64(txAdded))
}

func updateContractNotiPK(sqlTx *sql.Tx, pk uint) error {
	return updateCounter(sqlTx, "`contract_noti_pk`", int64(pk))
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

func updateCounterDelta(sqlTx *sql.Tx, field string, delta int64) error {
	if delta == 0 {
		return nil
	}

	query := []string{
		"UPDATE `counter`",
		fmt.Sprintf("SET %s = %s + %d", field, field, delta),
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
